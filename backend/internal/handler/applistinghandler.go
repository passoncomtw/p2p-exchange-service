package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"p2p-exchange/internal/logic"
	"p2p-exchange/internal/response"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

func AppCreateListingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateListingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		uidRaw, _ := r.Context().Value("sub").(json.Number)
		uid, _ := uidRaw.Int64()

		l := logic.NewAppCreateListingLogic(r.Context(), svcCtx)
		resp, err := l.Create(uid, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}

func AppListListingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListListingsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewAppListListingsLogic(r.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}

func AppGetListingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID int64 `path:"id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewAppGetListingLogic(r.Context(), svcCtx)
		resp, err := l.Get(req.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}

func AppCancelListingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelListingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		uidRaw, _ := r.Context().Value("sub").(json.Number)
		uid, _ := uidRaw.Int64()

		l := logic.NewAppCancelListingLogic(r.Context(), svcCtx)
		err := l.Cancel(uid, req.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(map[string]bool{"ok": true}))
		}
	}
}
