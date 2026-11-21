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

// Create 新增收款帳戶。
//
//	@Summary		新增收款帳戶
//	@Tags			app-payment-methods
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	app_interface.CreatePaymentMethodRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=app_interface.CreatePaymentMethodResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/payment-methods [post]
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

// List 查詢收款帳戶列表。
//
//	@Summary		查詢收款帳戶列表
//	@Tags			app-payment-methods
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Success		200	{object}	response.Body{data=app_interface.ListPaymentMethodsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/payment-methods [get]
func (h *PaymentMethodHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := ctxUID(r)
	resp, err := h.svc.List(r.Context(), uid)
	if err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// Delete 刪除收款帳戶。
//
//	@Summary		刪除收款帳戶
//	@Tags			app-payment-methods
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		int	true	"收款帳戶 ID"	Example(1)
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/payment-methods/{id} [delete]
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
