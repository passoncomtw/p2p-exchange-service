package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	cryptowithdraw_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/crypto_withdraw"
	"p2p-exchange/internal/response"
)

type CryptoWithdrawHandler struct {
	svc cryptowithdraw_service.CryptoWithdrawService
}

func NewCryptoWithdrawHandler(svc cryptowithdraw_service.CryptoWithdrawService) *CryptoWithdrawHandler {
	return &CryptoWithdrawHandler{svc: svc}
}

// Create 建立 USDT 提領申請。
//
//	@Summary		建立 USDT 提領申請
//	@Description	申請時會凍結對應餘額；最低 10 USDT。
//	@Tags			app-crypto-withdraw
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	app_interface.CryptoWithdrawRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=app_interface.CryptoWithdrawResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/crypto/withdraw [post]
func (h *CryptoWithdrawHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req app_interface.CryptoWithdrawRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 提領歸屬人一律取自 JWT context，不接受 request body 指定。
	uid := ctxUID(r)
	resp, err := h.svc.RequestWithdraw(r.Context(), uid, req.ToAddress, req.Amount)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// List 查詢 USDT 提領紀錄。
//
//	@Summary		查詢 USDT 提領紀錄
//	@Tags			app-crypto-withdraw
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	app_interface.AppListCryptoWithdrawalsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.AppListCryptoWithdrawalsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/crypto/withdrawals [get]
func (h *CryptoWithdrawHandler) List(w http.ResponseWriter, r *http.Request) {
	var req app_interface.AppListCryptoWithdrawalsRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	resp, err := h.svc.ListWithdrawals(r.Context(), uid, req.Limit, req.Offset)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}
