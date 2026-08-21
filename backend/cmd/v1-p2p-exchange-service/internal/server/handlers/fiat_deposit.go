package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	fiatdeposit_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/fiat_deposit"
	"p2p-exchange/internal/response"
)

type FiatDepositHandler struct {
	svc fiatdeposit_service.FiatDepositService
}

func NewFiatDepositHandler(svc fiatdeposit_service.FiatDepositService) *FiatDepositHandler {
	return &FiatDepositHandler{svc: svc}
}

func (h *FiatDepositHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req app_interface.FiatDepositRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 入金歸屬人一律取自 JWT context，不接受 request body 指定。
	uid := ctxUID(r)
	resp, err := h.svc.CreateDeposit(r.Context(), uid, req.Amount)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *FiatDepositHandler) List(w http.ResponseWriter, r *http.Request) {
	var req app_interface.AppListFiatDepositsRequest
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
