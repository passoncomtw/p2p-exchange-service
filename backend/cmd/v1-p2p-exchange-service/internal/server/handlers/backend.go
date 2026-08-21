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

func (h *BackendHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	username := ctxStringClaim(r, "username")
	role := ctxStringClaim(r, "role")
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(backend_interface.DashboardResponse{
		Username: username,
		Role:     role,
	}))
}

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

func (h *BackendHandler) PlatformWalletInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := h.adminSvc.GetPlatformWalletInfo(r.Context())
	if err != nil {
		code := appErrCode(err)
		httpx.WriteJsonCtx(r.Context(), w, code, response.Fail(code, err.Error()))
		return
	}
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(resp))
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
