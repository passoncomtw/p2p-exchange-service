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

// Create 建立 TWD 提領申請。
//
//	@Summary		建立 TWD 提領申請
//	@Description	狀態固定為 pending，等待後台審核；申請時會凍結對應餘額。
//	@Tags			app-fiat-withdraw
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	app_interface.FiatWithdrawRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=app_interface.FiatWithdrawResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/fiat/withdraw [post]
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

// List 查詢 TWD 提領紀錄。
//
//	@Summary		查詢 TWD 提領紀錄
//	@Tags			app-fiat-withdraw
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	app_interface.AppListFiatWithdrawalsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.AppListFiatWithdrawalsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/app/wallets/fiat/withdrawals [get]
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
