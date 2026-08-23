package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"p2p-exchange/internal/response"
)

// TestProfileHandleReadsUsernameClaimDirectly 鎖住 go-zero JWT middleware 的實際行為：
// 每個 claim 是個別以 context.WithValue(ctx, k, v) 寫入 context，沒有 "payload" 包裝層。
// 舊實作讀 Value("payload").(map[string]interface{}) 永遠斷言失敗，回傳空字串 username。
func TestProfileHandleReadsUsernameClaimDirectly(t *testing.T) {
	h := NewProfileHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/app/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), "username", "demo_user"))
	rec := httptest.NewRecorder()

	h.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body response.Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %#v", body.Data)
	}
	if got := data["username"]; got != "demo_user" {
		t.Errorf("username = %v, want %q", got, "demo_user")
	}
}

// TestProfileHandleMissingClaimReturnsEmptyUsername 沒有 username claim 時（理論上不會發生，
// 因為路由本身受 JWT 保護）回傳空字串而不是 panic。
func TestProfileHandleMissingClaimReturnsEmptyUsername(t *testing.T) {
	h := NewProfileHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/app/profile", nil)
	rec := httptest.NewRecorder()

	h.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
