package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	v1_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/v1"
	v1_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/v1"
	"p2p-exchange/internal/response"
)

type V1Handler struct {
	svc v1_service.V1Service
}

func NewV1Handler(svc v1_service.V1Service) *V1Handler {
	return &V1Handler{svc: svc}
}

func (h *V1Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req v1_interface.V1CreateOrderRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.svc.Create(r.Context(), req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *V1Handler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListMine(r.Context())
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *V1Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	var req v1_interface.V1OrderIDRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if err := h.svc.Cancel(r.Context(), req.ID); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(nil))
}

func (h *V1Handler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	var req v1_interface.V1AdminListOrdersRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.svc.AdminList(r.Context(), req.Status)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *V1Handler) AdminGetOrder(w http.ResponseWriter, r *http.Request) {
	var req v1_interface.V1OrderIDRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.svc.AdminGet(r.Context(), req.ID)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

func (h *V1Handler) AdminCompleteOrder(w http.ResponseWriter, r *http.Request) {
	var req v1_interface.V1OrderIDRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if err := h.svc.AdminComplete(r.Context(), req.ID); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(nil))
}
