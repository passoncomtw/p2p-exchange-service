package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"p2p-exchange/internal/logic"
	"p2p-exchange/internal/response"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

func BackendListOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BackendListOrdersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewBackendListOrdersLogic(r.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}

func BackendResolveOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResolveOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewBackendResolveOrderLogic(r.Context(), svcCtx)
		resp, err := l.Resolve(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}
