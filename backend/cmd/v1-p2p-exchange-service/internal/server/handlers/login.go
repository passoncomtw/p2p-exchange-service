package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	auth_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"
	"p2p-exchange/internal/response"
)

type LoginHandler struct {
	authSvc auth_service.AuthService
}

func NewLoginHandler(authSvc auth_service.AuthService) *LoginHandler {
	return &LoginHandler{authSvc: authSvc}
}

// Handle 處理 App 使用者登入。
//
//	@Summary		App 使用者登入
//	@Description	以帳號密碼登入，成功回傳 JWT 與使用者資訊。
//	@Tags			app-auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		app_interface.LoginRequest	true	"登入資訊"
//	@Success		200		{object}	response.Body{data=app_interface.AppLoginResponse}
//	@Failure		400		{object}	response.ErrorBody	"請求格式錯誤"
//	@Failure		401		{object}	response.ErrorBody	"帳號或密碼錯誤"
//	@Router			/app/auth/login [post]
func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req app_interface.LoginRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	out, err := h.authSvc.Login(r.Context(), auth_service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, err.Error()))
		return
	}

	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(app_interface.AppLoginResponse{
		AccessToken: out.AccessToken,
		ExpireIn:    out.ExpireIn,
		User: app_interface.AppLoginUserInfo{
			ID:      out.UserID,
			Account: out.Username,
			Name:    out.Username,
		},
	}))
}
