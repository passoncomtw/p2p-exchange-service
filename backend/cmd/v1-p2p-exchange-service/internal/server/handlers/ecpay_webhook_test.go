package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// stubWebhookService 記錄收到的 params，用來驗證 handler 是否把「所有」form 欄位都轉交給驗簽。
type stubWebhookService struct {
	got  map[string]string
	ok   bool
	msg  string
	call int
}

func (s *stubWebhookService) HandleNotify(_ context.Context, params map[string]string) (bool, string) {
	s.call++
	s.got = params
	return s.ok, s.msg
}

func postForm(t *testing.T, h *ECPayWebhookHandler, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhook/ecpay/notify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.Notify(rec, req)
	return rec
}

// TestNotifyPassesAllFormFields handler 不得白名單擷取欄位：
// CheckMacValue 是對整組參數計算的，少帶（或多帶）任何欄位都會讓驗簽失敗。
func TestNotifyPassesAllFormFields(t *testing.T) {
	svc := &stubWebhookService{ok: true, msg: "OK"}
	h := NewECPayWebhookHandler(svc)

	form := url.Values{}
	form.Set("MerchantTradeNo", "P000123456789012")
	form.Set("RtnCode", "1")
	form.Set("TradeNo", "2608211500123456")
	form.Set("CheckMacValue", "ABC123")
	// ECPay 日後新增的未知欄位同樣必須進入驗簽。
	form.Set("SomeFutureField", "future-value")

	postForm(t, h, form)

	want := map[string]string{
		"MerchantTradeNo": "P000123456789012",
		"RtnCode":         "1",
		"TradeNo":         "2608211500123456",
		"CheckMacValue":   "ABC123",
		"SomeFutureField": "future-value",
	}
	if !reflect.DeepEqual(svc.got, want) {
		t.Fatalf("handler must forward every form field: got %v, want %v", svc.got, want)
	}
}

// TestNotifyResponseFormat ECPay 通知協定要求純文字 "1|OK" / "0|<message>"，且一律 HTTP 200。
// 回 4xx/5xx 或 JSON envelope 會讓 ECPay 判定通知未送達而無限重送。
func TestNotifyResponseFormat(t *testing.T) {
	tests := []struct {
		name     string
		ok       bool
		msg      string
		wantBody string
	}{
		{name: "成功", ok: true, msg: "OK", wantBody: "1|OK"},
		{name: "簽章錯誤", ok: false, msg: "signature mismatch", wantBody: "0|signature mismatch"},
		{name: "未設定 ECPay", ok: false, msg: "ECPay not configured", wantBody: "0|ECPay not configured"},
		{name: "查無記錄", ok: false, msg: "deposit not found", wantBody: "0|deposit not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewECPayWebhookHandler(&stubWebhookService{ok: tt.ok, msg: tt.msg})
			rec := postForm(t, h, url.Values{"RtnCode": {"1"}})

			if rec.Code != http.StatusOK {
				t.Fatalf("expected HTTP 200 for every outcome, got %d", rec.Code)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
				t.Fatalf("Content-Type = %q, want text/plain", ct)
			}
		})
	}
}

// TestNotifyParseErrorDoesNotReachService form 解析失敗時不得呼叫 service，並回 0|parse error。
func TestNotifyParseErrorDoesNotReachService(t *testing.T) {
	svc := &stubWebhookService{ok: true, msg: "OK"}
	h := NewECPayWebhookHandler(svc)

	// 非法的 percent-encoding 會讓 ParseForm 失敗。
	req := httptest.NewRequest(http.MethodPost, "/webhook/ecpay/notify", strings.NewReader("RtnCode=%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.Notify(rec, req)

	if svc.call != 0 {
		t.Fatalf("service must not be called when form parsing fails, got %d calls", svc.call)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != "0|parse error" {
		t.Fatalf("got %d %q, want 200 %q", rec.Code, rec.Body.String(), "0|parse error")
	}
}
