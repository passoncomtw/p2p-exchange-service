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

// Create 建立訂單（接單）。
//
//	@Summary		建立訂單（接單）
//	@Description	不可接自己的掛單；換算法幣金額須落在掛單的 minOrderFiat ~ maxOrderFiat。
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	app_interface.CreateOrderRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=app_interface.CreateOrderResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"掛單不存在"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/orders [post]
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

// List 查詢目前使用者的訂單列表。
//
//	@Summary		查詢訂單列表
//	@Description	role 為 buyer / seller 時只回該身分的訂單；留空則買賣雙方皆含。
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query		app_interface.ListOrdersRequest	false	"查詢條件"
//	@Success		200		{object}	response.Body{data=app_interface.ListOrdersResponse}
//	@Failure		400		{object}	response.ErrorBody
//	@Router			/app/orders [get]
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

// Get 查詢訂單詳情。
//
//	@Summary		查詢訂單詳情
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		int	true	"訂單 ID"	Example(1)
//	@Success		200	{object}	response.Body{data=app_interface.OrderItem}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"訂單不存在"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/orders/{id} [get]
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

// Pay 標記已付款。
//
//	@Summary		標記已付款
//	@Description	訂單必須為 matched 狀態，成功後轉為 paid。
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		int	true	"訂單 ID"	Example(1)
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"訂單不存在"
//	@Failure		409	{object}	response.ErrorBody	"訂單狀態已變更"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/orders/{id}/pay [put]
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

// Confirm 確認收款。
//
//	@Summary		確認收款
//	@Description	訂單必須為 paid 狀態，成功後轉為 completed 並放行幣給買方。
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		int	true	"訂單 ID"	Example(1)
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"訂單不存在"
//	@Failure		409	{object}	response.ErrorBody	"訂單狀態已變更"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/orders/{id}/confirm [put]
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

// Cancel 取消訂單。
//
//	@Summary		取消訂單
//	@Description	僅 matched 狀態可取消。
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"訂單 ID"	Example(1)
//	@Param			request	body	app_interface.CancelOrderRequest	true	"請求內容"
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"訂單不存在"
//	@Failure		409	{object}	response.ErrorBody	"訂單狀態已變更"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/orders/{id}/cancel [put]
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

// Dispute 提出申訴。
//
//	@Summary		提出申訴
//	@Description	僅 paid 狀態可申訴，成功後轉為 disputed。
//	@Tags			app-orders
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"訂單 ID"	Example(1)
//	@Param			request	body	app_interface.DisputeOrderRequest	true	"請求內容"
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"訂單不存在"
//	@Failure		409	{object}	response.ErrorBody	"訂單狀態已變更"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/orders/{id}/dispute [put]
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
