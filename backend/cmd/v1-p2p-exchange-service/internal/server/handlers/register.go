package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	auth_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"
	"p2p-exchange/internal/response"
)

type RegisterHandler struct {
	authSvc auth_service.AuthService
}

func NewRegisterHandler(authSvc auth_service.AuthService) *RegisterHandler {
	return &RegisterHandler{authSvc: authSvc}
}

// Handle 處理 POST /app/auth/register。
// 註冊成功等同登入成功，回傳與 /app/auth/login 相同的 AppLoginResponse。
// 錯誤狀態碼由 service 回傳的 AppError 決定：400（格式不符）/ 409（帳號已存在）/ 500（其他）。
//
//	@Summary		App 使用者註冊
//	@Description	成功等同登入，回傳與登入相同的結構。
//	@Tags			app-auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	app_interface.RegisterRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=app_interface.AppLoginResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		409	{object}	response.ErrorBody	"帳號已存在"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/auth/register [post]
func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req app_interface.RegisterRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	out, err := h.authSvc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
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
