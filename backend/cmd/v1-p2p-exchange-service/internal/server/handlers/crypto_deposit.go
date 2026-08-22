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
