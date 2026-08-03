package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	payment_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/payment_method"
	"p2p-exchange/internal/response"
)

type PaymentMethodHandler struct {
	svc payment_service.PaymentMethodService
}

func NewPaymentMethodHandler(svc payment_service.PaymentMethodService) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc}
}

func (h *PaymentMethodHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req app_interface.CreatePaymentMethodRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	resp, err := h.svc.Create(r.Context(), uid, req)
	if err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *PaymentMethodHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := ctxUID(r)
	resp, err := h.svc.List(r.Context(), uid)
	if err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *PaymentMethodHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `path:"id"`
	}
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	if err := h.svc.Delete(r.Context(), uid, req.ID); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}
