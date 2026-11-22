package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	cryptodeposit_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/crypto_deposit"
	"p2p-exchange/internal/response"
)

type CryptoDepositHandler struct {
	svc cryptodeposit_service.CryptoDepositService
}

func NewCryptoDepositHandler(svc cryptodeposit_service.CryptoDepositService) *CryptoDepositHandler {
	return &CryptoDepositHandler{svc: svc}
}

// GetInfo 取得 USDT 充值地址與 memo。
//
//	@Summary		取得 USDT 充值地址與 memo
//	@Tags			app-crypto-deposit
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Success		200	{object}	response.Body{data=app_interface.GetCryptoDepositInfoResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		503	{object}	response.ErrorBody	"USDT 充值功能尚未開放"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/crypto/deposit-info [get]
func (h *CryptoDepositHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	// memo 綁定 JWT 中的 uid，不接受呼叫端指定，避免充值被導到他人帳戶。
	uid := ctxUID(r)
	resp, err := h.svc.GetDepositInfo(r.Context(), uid)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// List 查詢 USDT 充值紀錄。
//
//	@Summary		查詢 USDT 充值紀錄
//	@Tags			app-crypto-deposit
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	app_interface.AppListCryptoDepositsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.AppListCryptoDepositsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/crypto/deposits [get]
func (h *CryptoDepositHandler) List(w http.ResponseWriter, r *http.Request) {
	var req app_interface.AppListCryptoDepositsRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	resp, err := h.svc.ListDeposits(r.Context(), uid, req.Limit, req.Offset)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}
