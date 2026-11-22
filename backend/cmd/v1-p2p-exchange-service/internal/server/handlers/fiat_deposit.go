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

// Create 建立 TWD 入金訂單（導轉 ECPay）。
//
//	@Summary		建立 TWD 入金訂單（導轉 ECPay）
//	@Description	前端以 POST form 將 formParams 送往 paymentUrl。金額 100 ~ 50000 TWD。
//	@Tags			app-fiat-deposit
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	app_interface.FiatDepositRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=app_interface.FiatDepositResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		503	{object}	response.ErrorBody	"ECPay 未啟用"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/fiat/deposit [post]
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

// List 查詢 TWD 入金紀錄。
//
//	@Summary		查詢 TWD 入金紀錄
//	@Tags			app-fiat-deposit
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	app_interface.AppListFiatDepositsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.AppListFiatDepositsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/fiat/deposits [get]
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
