package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/rest/pathvar"

	backend_admin_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_admin"
)

// stubAdminService 記錄 handler 實際傳給 service 的審核人與申請 ID。
// 其餘方法沿用內嵌的 nil 介面：走到未預期路徑會直接 panic，不會靜默通過。
type stubAdminService struct {
	backend_admin_service.BackendAdminService

	calls          int
	gotReviewerID  int64
	gotWithdrawID  int64
	gotAction      string
	gotReasonValue string
}

func (s *stubAdminService) ReviewFiatWithdrawal(_ context.Context, reviewerID, id int64, action, reason string) error {
	s.calls++
	s.gotReviewerID = reviewerID
	s.gotWithdrawID = id
	s.gotAction = action
	s.gotReasonValue = reason
	return nil
}

// TestReviewFiatWithdrawalTakesReviewerFromJWTOnly 審核人只能來自 Backend JWT context。
//
// 攻擊情境：呼叫端在 body 塞入 reviewerId / adminUid / reviewed_by 想把稽核紀錄
// （fiat_withdrawals.reviewed_by）寫成別的管理員，藉此嫁禍或掩蓋是誰核可了這筆撥款。
// BackendReviewFiatWithdrawalRequest 根本沒有這類欄位，handler 一律取 ctxUID(r)，
// 因此 body 帶什麼都不會改變寫入的審核人。
func TestReviewFiatWithdrawalTakesReviewerFromJWTOnly(t *testing.T) {
	svc := &stubAdminService{}
	h := NewBackendHandler(nil, svc)

	body := `{"action":"approve","reason":"ok","reviewerId":999,"adminUid":999,"reviewed_by":999,"uid":999}`
	req := httptest.NewRequest(http.MethodPut, "/backend/fiat-withdrawals/42/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = pathvar.WithVars(req, map[string]string{"id": "42"})
	// go-zero 的 JWT middleware 會把 claims 放進 context；uid 為 json.Number。
	req = req.WithContext(context.WithValue(req.Context(), "uid", json.Number("7")))

	rec := httptest.NewRecorder()
	h.ReviewFiatWithdrawal(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if svc.calls != 1 {
		t.Fatalf("expected exactly 1 service call, got %d", svc.calls)
	}
	if svc.gotReviewerID != 7 {
		t.Fatalf("reviewer must come from JWT context (7), got %d — request body must never win", svc.gotReviewerID)
	}
	if svc.gotWithdrawID != 42 {
		t.Fatalf("expected withdrawal id 42 from path, got %d", svc.gotWithdrawID)
	}
	if svc.gotAction != "approve" {
		t.Fatalf("expected action approve, got %q", svc.gotAction)
	}
}

// TestReviewFiatWithdrawalWithoutJWTUIDPassesZero 沒有 JWT uid 時審核人為 0，
// 仍然不會採用 body 傳入的值（此路由由 Backend JWT middleware 保護，正常情況不會發生）。
func TestReviewFiatWithdrawalWithoutJWTUIDPassesZero(t *testing.T) {
	svc := &stubAdminService{}
	h := NewBackendHandler(nil, svc)

	body := `{"action":"reject","reason":"帳號有誤","reviewerId":999}`
	req := httptest.NewRequest(http.MethodPut, "/backend/fiat-withdrawals/42/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = pathvar.WithVars(req, map[string]string{"id": "42"})

	rec := httptest.NewRecorder()
	h.ReviewFiatWithdrawal(rec, req)

	if svc.gotReviewerID != 0 {
		t.Fatalf("expected reviewer 0 without JWT uid, got %d", svc.gotReviewerID)
	}
}
