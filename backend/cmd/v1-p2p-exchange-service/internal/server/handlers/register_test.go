package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"
	apierrors "p2p-exchange/internal/errors"
)

// stubAuthService 只實作 Register；其餘方法沿用內嵌的 nil 介面，
// 走到未預期路徑會直接 panic，不會靜默通過。
type stubAuthService struct {
	auth_service.AuthService

	calls       int
	gotUsername string
	gotPassword string
	out         *auth_service.LoginOutput
	err         error
}

func (s *stubAuthService) Register(_ context.Context, username, password string) (*auth_service.LoginOutput, error) {
	s.calls++
	s.gotUsername = username
	s.gotPassword = password
	if s.err != nil {
		return nil, s.err
	}
	return s.out, nil
}

func doRegister(t *testing.T, svc *stubAuthService, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewRegisterHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/app/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Handle(rec, req)
	return rec
}

// TestRegisterHandlerMapsErrorCodes 註冊錯誤必須依語意分流，不可全部包成同一個狀態碼。
func TestRegisterHandlerMapsErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{name: "格式錯誤", err: apierrors.New(400, "username must be between 3 and 50 characters"), wantCode: http.StatusBadRequest},
		{name: "帳號已存在", err: apierrors.ErrUserAlreadyExists, wantCode: http.StatusConflict},
		{name: "內部錯誤", err: apierrors.ErrInternal, wantCode: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRegister(t, &stubAuthService{err: tt.err}, `{"username":"alice","password":"12345678"}`)
			if rec.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, rec.Code, rec.Body.String())
			}
			var body struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Data    any    `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("回應非合法 JSON：%v", err)
			}
			if body.Code != tt.wantCode {
				t.Fatalf("body code 應為 %d，實際為 %d", tt.wantCode, body.Code)
			}
			if body.Data != nil {
				t.Fatalf("錯誤回應不應帶 data：%v", body.Data)
			}
		})
	}
}

// TestRegisterHandlerNeverEchoesPassword 任何回應內容都不得夾帶送出的密碼。
func TestRegisterHandlerNeverEchoesPassword(t *testing.T) {
	const password = "sup3r-secret-pw"
	svc := &stubAuthService{err: apierrors.ErrUserAlreadyExists}
	rec := doRegister(t, svc, `{"username":"alice","password":"`+password+`"}`)

	if strings.Contains(rec.Body.String(), password) {
		t.Fatalf("回應夾帶了密碼：%s", rec.Body.String())
	}
	if svc.gotPassword != password {
		t.Fatalf("handler 應原封不動把密碼交給 service，實際為 %q", svc.gotPassword)
	}
}

// TestRegisterHandlerReturnsLoginShapedResponse 註冊成功回傳與登入相同格式的回應。
func TestRegisterHandlerReturnsLoginShapedResponse(t *testing.T) {
	svc := &stubAuthService{out: &auth_service.LoginOutput{
		AccessToken: "token-abc",
		ExpireIn:    3600,
		UserID:      4242,
		Username:    "alice",
	}}
	rec := doRegister(t, svc, `{"username":"  alice  ","password":"12345678"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
			ExpireIn    int64  `json:"expireIn"`
			User        struct {
				ID      int64  `json:"id"`
				Account string `json:"account"`
				Name    string `json:"name"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應非合法 JSON：%v", err)
	}
	if body.Code != 0 {
		t.Fatalf("expected code 0, got %d", body.Code)
	}
	if body.Data.AccessToken != "token-abc" || body.Data.ExpireIn != 3600 {
		t.Fatalf("unexpected data: %+v", body.Data)
	}
	if body.Data.User.ID != 4242 || body.Data.User.Account != "alice" || body.Data.User.Name != "alice" {
		t.Fatalf("unexpected user: %+v", body.Data.User)
	}
	// 去空白由 service 負責，handler 原封不動轉交。
	if svc.gotUsername != "  alice  " {
		t.Fatalf("handler 不應改寫帳號，實際傳入 %q", svc.gotUsername)
	}
}
