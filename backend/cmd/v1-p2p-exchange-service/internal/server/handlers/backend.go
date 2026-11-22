package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	backend_admin_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_admin"
	backend_auth_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_auth"
	"p2p-exchange/internal/response"
)

type BackendHandler struct {
	authSvc  backend_auth_service.BackendAuthService
	adminSvc backend_admin_service.BackendAdminService
}

func NewBackendHandler(authSvc backend_auth_service.BackendAuthService, adminSvc backend_admin_service.BackendAdminService) *BackendHandler {
	return &BackendHandler{authSvc: authSvc, adminSvc: adminSvc}
}

// Login 後台管理員登入。
//
//	@Summary		後台管理員登入
//	@Description	預設帳號 admin001。
//	@Tags			backend-auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	backend_interface.LoginRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=backend_interface.LoginResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		401	{object}	response.ErrorBody	"帳號或密碼錯誤"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/auth/login [post]
func (h *BackendHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.LoginRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	out, err := h.authSvc.Login(r.Context(), backend_auth_service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(backend_interface.LoginResponse{
		Token:     out.Token,
		ExpiresIn: out.ExpiresIn,
	}))
}

// Dashboard 後台 Dashboard。
//
//	@Summary		後台 Dashboard
//	@Tags			backend-admin
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Success		200	{object}	response.Body{data=backend_interface.DashboardResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/dashboard [get]
func (h *BackendHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	username := ctxStringClaim(r, "username")
	role := ctxStringClaim(r, "role")
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(backend_interface.DashboardResponse{
		Username: username,
		Role:     role,
	}))
}

// ListListings 後台查詢掛單列表。
//
//	@Summary		後台查詢掛單列表
//	@Tags			backend-admin
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	backend_interface.BackendListListingsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.ListListingsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/listings [get]
func (h *BackendHandler) ListListings(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.BackendListListingsRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.adminSvc.ListListings(r.Context(), req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// ListOrders 後台查詢訂單列表。
//
//	@Summary		後台查詢訂單列表
//	@Tags			backend-admin
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	backend_interface.BackendListOrdersRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=app_interface.ListOrdersResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/orders [get]
func (h *BackendHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.BackendListOrdersRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.adminSvc.ListOrders(r.Context(), req)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// ResolveOrder 後台處理爭議訂單。
//
//	@Summary		後台處理爭議訂單
//	@Description	僅 disputed 狀態可處理。action=complete 放行給買方；action=refund 退回賣方並將訂單轉為 refunded。
//	@Tags			backend-admin
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"訂單 ID"	Example(1)
//	@Param			request	body	backend_interface.ResolveOrderRequest	true	"請求內容"
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"訂單不存在"
//	@Failure		409	{object}	response.ErrorBody	"訂單狀態已變更"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/orders/{id}/resolve [put]
func (h *BackendHandler) ResolveOrder(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.ResolveOrderRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if err := h.adminSvc.ResolveOrder(r.Context(), req); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(nil))
}

// ListMembers 後台查詢會員列表。
//
//	@Summary		後台查詢會員列表
//	@Tags			backend-members
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	backend_interface.BackendListMembersRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=backend_interface.BackendListMembersResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/members [get]
func (h *BackendHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.BackendListMembersRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.adminSvc.ListMembers(r.Context(), req.Keyword, req.Limit, req.Offset)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// Deposit 後台手動為會員入金。
//
//	@Summary		後台手動為會員入金
//	@Description	操作人一律取自 Backend JWT，不接受 body 指定。
//	@Tags			backend-members
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"會員 ID"	Example(1)
//	@Param			request	body	backend_interface.BackendDepositRequest	true	"請求內容"
//	@Success		200	{object}	response.Body{data=backend_interface.BackendDepositResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"會員不存在"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/members/{id}/deposit [post]
func (h *BackendHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.BackendDepositRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 操作人一律取自 Backend JWT context，不從 request body 讀取，
	// 否則稽核記錄可被呼叫端偽造成任意管理員。
	adminUID := ctxUID(r)
	resp, err := h.adminSvc.DepositToMember(r.Context(), adminUID, req.ID, req.Currency, req.Amount)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// PlatformWalletInfo 查詢平台熱錢包鏈上餘額與虛擬餘額。
//
//	@Summary		查詢平台熱錢包鏈上餘額與虛擬餘額
//	@Tags			backend-platform-wallet
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Success		200	{object}	response.Body{data=backend_interface.PlatformWalletInfoResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/platform/wallet-info [get]
func (h *BackendHandler) PlatformWalletInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := h.adminSvc.GetPlatformWalletInfo(r.Context())
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// ListFiatWithdrawals 後台查詢法幣提領申請列表。
//
//	@Summary		後台查詢法幣提領申請列表
//	@Description	status 預設 pending；傳 all 回全部狀態。
//	@Tags			backend-fiat-withdrawals
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			request	query	backend_interface.BackendListFiatWithdrawalsRequest	false	"查詢條件"
//	@Success		200	{object}	response.Body{data=backend_interface.BackendListFiatWithdrawalsResponse}
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/fiat-withdrawals [get]
func (h *BackendHandler) ListFiatWithdrawals(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.BackendListFiatWithdrawalsRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	resp, err := h.adminSvc.ListFiatWithdrawals(r.Context(), req.Status, req.Limit, req.Offset)
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
}

// ReviewFiatWithdrawal 後台審核法幣提領申請（核可/駁回）。
//
//	@Summary		後台審核法幣提領申請（核可/駁回）
//	@Description	reject 時必須填寫 reason；審核人一律取自 Backend JWT。
//	@Tags			backend-fiat-withdrawals
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"提領申請 ID"	Example(1)
//	@Param			request	body	backend_interface.BackendReviewFiatWithdrawalRequest	true	"請求內容"
//	@Success		200	{object}	response.Body
//	@Failure		400	{object}	response.ErrorBody	"請求格式錯誤或驗證失敗"
//	@Failure		404	{object}	response.ErrorBody	"申請不存在"
//	@Failure		409	{object}	response.ErrorBody	"此申請已審核完畢"
//	@Failure		500	{object}	response.ErrorBody	"內部錯誤"
//	@Router			/backend/fiat-withdrawals/{id}/review [put]
func (h *BackendHandler) ReviewFiatWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req backend_interface.BackendReviewFiatWithdrawalRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 審核人一律取自 Backend JWT context，不從 request body 讀取，
	// 否則稽核記錄（fiat_withdrawals.reviewed_by）可被呼叫端偽造成任意管理員。
	reviewerID := ctxUID(r)
	if err := h.adminSvc.ReviewFiatWithdrawal(r.Context(), reviewerID, req.ID, req.Action, req.Reason); err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(nil))
}

func ctxStringClaim(r *http.Request, key string) string {
	v := r.Context().Value(key)
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	default:
		return ""
	}
}
