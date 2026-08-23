package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
)

// stubPushTokenRepo 記錄 handler 實際傳給 repository 的 uid 與 token。
// 其餘方法沿用內嵌的 nil 介面：走到未預期路徑會直接 panic，不會靜默通過。
type stubPushTokenRepo struct {
	userrepo.UserRepository

	calls    int
	gotUID   int64
	gotToken string
	err      error
}

func (r *stubPushTokenRepo) UpdatePushToken(_ context.Context, userID int64, token string) error {
	r.calls++
	r.gotUID = userID
	r.gotToken = token
	return r.err
}

func doUpdatePushToken(t *testing.T, repo *stubPushTokenRepo, body string, uid json.Number) *httptest.ResponseRecorder {
	t.Helper()
	h := NewProfileHandler(repo)
	req := httptest.NewRequest(http.MethodPut, "/app/profile/push-token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// go-zero 的 JWT middleware 會把 claims 放進 context；uid 為 json.Number。
	req = req.WithContext(context.WithValue(req.Context(), "uid", uid))
	rec := httptest.NewRecorder()
	h.UpdatePushToken(rec, req)
	return rec
}

// TestUpdatePushTokenSkipsEmptyToken token 為空字串時直接回成功，不得寫入 DB。
// 這是 legacy 的既有行為（清空 token 尚未支援），不可「順手」改成允許清空。
func TestUpdatePushTokenSkipsEmptyToken(t *testing.T) {
	repo := &stubPushTokenRepo{}
	rec := doUpdatePushToken(t, repo, `{"token":""}`, json.Number("7"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.calls != 0 {
		t.Fatalf("空 token 仍寫入 DB（%d 次，值 %q）", repo.calls, repo.gotToken)
	}
}

// TestUpdatePushTokenRejectsMissingField token 欄位未帶時由 httpx.Parse 擋成 400，且不寫入 DB。
// 與 legacy 的 RegisterPushTokenRequest 一致（`json:"token"` 未標 optional 即為必填）。
func TestUpdatePushTokenRejectsMissingField(t *testing.T) {
	repo := &stubPushTokenRepo{}
	rec := doUpdatePushToken(t, repo, `{}`, json.Number("7"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.calls != 0 {
		t.Fatalf("解析失敗仍寫入 DB（%d 次）", repo.calls)
	}
}

// TestUpdatePushTokenWritesNonEmptyToken 非空 token 必須原封不動寫入。
func TestUpdatePushTokenWritesNonEmptyToken(t *testing.T) {
	repo := &stubPushTokenRepo{}
	rec := doUpdatePushToken(t, repo, `{"token":"ExponentPushToken[abc123]"}`, json.Number("7"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.calls != 1 {
		t.Fatalf("expected 1 UpdatePushToken call, got %d", repo.calls)
	}
	if repo.gotToken != "ExponentPushToken[abc123]" {
		t.Fatalf("unexpected token: %q", repo.gotToken)
	}
}

// TestUpdatePushTokenTakesUIDFromJWTOnly 更新對象只能來自 JWT context。
//
// 攻擊情境：呼叫端在 body 塞 uid / user_id 想把自己的推播 token 寫到別人的帳號上
// （之後該使用者的訂單通知就會推到攻擊者的裝置）。RegisterPushTokenRequest 沒有這類欄位，
// handler 一律取 ctxUID(r)，因此 body 帶什麼都不會改變寫入對象。
func TestUpdatePushTokenTakesUIDFromJWTOnly(t *testing.T) {
	repo := &stubPushTokenRepo{}
	body := `{"token":"ExponentPushToken[attacker]","uid":999,"user_id":999,"userId":999}`
	rec := doUpdatePushToken(t, repo, body, json.Number("7"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.gotUID != 7 {
		t.Fatalf("寫入對象應為 JWT 的 uid=7，實際為 %d", repo.gotUID)
	}
}

// TestUpdatePushTokenDoesNotLeakDBError repository 錯誤只回泛用 500，不外露 DB 錯誤字串。
func TestUpdatePushTokenDoesNotLeakDBError(t *testing.T) {
	repo := &stubPushTokenRepo{err: errors.New(`pq: relation "app_users" does not exist on host db-prod-1`)}
	rec := doUpdatePushToken(t, repo, `{"token":"ExponentPushToken[abc123]"}`, json.Number("7"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "db-prod-1") || strings.Contains(rec.Body.String(), "app_users") {
		t.Fatalf("內部 DB 錯誤字串外洩：%s", rec.Body.String())
	}
}
