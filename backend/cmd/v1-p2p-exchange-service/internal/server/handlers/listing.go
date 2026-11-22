package handlers

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	listing_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/listing"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/response"
)

type ListingHandler struct {
	svc listing_service.ListingService
}

func NewListingHandler(svc listing_service.ListingService) *ListingHandler {
	return &ListingHandler{svc: svc}
}

// Create 建立掛單。
//
//	@Summary		建立掛單
//	@Description	sell 掛單必須帶 paymentMethodId，並會凍結 totalAmount 的幣；buy 掛單不凍結。
//	@Tags			app-listings
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		app_interface.CreateListingRequest	true	"掛單內容"
//	@Success		200		{object}	response.Body{data=app_interface.CreateListingResponse}
//	@Failure		400		{object}	response.ErrorBody	"參數驗證失敗或餘額不足"
//	@Router			/app/listings [post]
func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req app_interface.CreateListingRequest
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

// List 查詢掛單列表。
//
//	@Summary		查詢掛單列表
//	@Tags			app-listings
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	app_interface.ListListingsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.ListListingsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/listings [get]
func (h *ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	var req app_interface.ListListingsRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.svc.List(r.Context(), req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// ListMine 查詢我的掛單。
//
//	@Summary		查詢我的掛單
//	@Tags			app-listings
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	app_interface.ListListingsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.ListListingsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/listings/mine [get]
func (h *ListingHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	var req app_interface.ListListingsRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	resp, err := h.svc.ListMine(r.Context(), uid, req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// Get 查詢掛單詳情。
//
//	@Summary		查詢掛單詳情
//	@Tags			app-listings
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		int	true	"掛單 ID"	Example(1)
//	@Success		200	{object}	response.Body{data=app_interface.ListingItem}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"掛單不存在"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/listings/{id} [get]
func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request) {
	var req app_interface.GetListingRequest
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

// Cancel 取消掛單。
//
//	@Summary		取消掛單
//	@Description	僅 active / paused 的掛單可取消；仍有進行中訂單時回 400；sell 掛單會解凍剩餘數量。
//	@Tags			app-listings
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		int	true	"掛單 ID"	Example(1)
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"掛單不存在"
//	@Failure		409	{object}	response.ErrorBody	"掛單狀態已變更"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/listings/{id}/cancel [put]
func (h *ListingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	var req app_interface.CancelListingRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	uid := ctxUID(r)
	if err := h.svc.Cancel(r.Context(), uid, req.ID); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}

func appErrCode(err error) int {
	var ae *apierrors.AppError
	if errors.As(err, &ae) {
		return ae.Code
	}
	return http.StatusInternalServerError
}
