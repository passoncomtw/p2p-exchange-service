package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	order_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/order"
	"p2p-exchange/internal/response"
)

type OrderHandler struct {
	svc order_service.OrderService
}

func NewOrderHandler(svc order_service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req app_interface.CreateOrderRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	resp, err := h.svc.Create(r.Context(), uid, req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	var req app_interface.ListOrdersRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	resp, err := h.svc.List(r.Context(), uid, req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	var req app_interface.GetOrderRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.svc.Get(r.Context(), req.ID)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	var req app_interface.UpdateOrderPathRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	if err := h.svc.Pay(r.Context(), uid, req.ID); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}

func (h *OrderHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var req app_interface.UpdateOrderPathRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	if err := h.svc.Confirm(r.Context(), uid, req.ID); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	var req app_interface.CancelOrderRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	req.ID = req.ID
	if err := h.svc.Cancel(r.Context(), uid, req); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}

func (h *OrderHandler) Dispute(w http.ResponseWriter, r *http.Request) {
	var req app_interface.DisputeOrderRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	if err := h.svc.Dispute(r.Context(), uid, req); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}
