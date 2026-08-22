package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	fiatwithdraw_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/fiat_withdraw"
	"p2p-exchange/internal/response"
)

type FiatWithdrawHandler struct {
	svc fiatwithdraw_service.FiatWithdrawService
}

func NewFiatWithdrawHandler(svc fiatwithdraw_service.FiatWithdrawService) *FiatWithdrawHandler {
	return &FiatWithdrawHandler{svc: svc}
}

func (h *FiatWithdrawHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req app_interface.FiatWithdrawRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 提領歸屬人一律取自 JWT context，不接受 request body 指定。
	uid := ctxUID(r)
	resp, err := h.svc.RequestWithdraw(r.Context(), uid, req.Amount, req.BankCode, req.BankAccount, req.AccountName)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *FiatWithdrawHandler) List(w http.ResponseWriter, r *http.Request) {
	var req app_interface.AppListFiatWithdrawalsRequest
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
