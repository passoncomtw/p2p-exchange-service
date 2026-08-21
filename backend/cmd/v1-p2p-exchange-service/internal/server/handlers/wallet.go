package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	wallet_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/wallet"
	"p2p-exchange/internal/response"
)

type WalletHandler struct {
	svc wallet_service.WalletService
}

func NewWalletHandler(svc wallet_service.WalletService) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := ctxUID(r)
	resp, err := h.svc.ListWallets(r.Context(), uid)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *WalletHandler) ListLedgers(w http.ResponseWriter, r *http.Request) {
	var req app_interface.ListWalletLedgersRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// uid 一律取自 JWT context，不接受呼叫端指定，避免查到他人帳本。
	uid := ctxUID(r)
	resp, err := h.svc.ListLedgers(r.Context(), uid, req.Currency, req.Limit, req.Offset)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}
