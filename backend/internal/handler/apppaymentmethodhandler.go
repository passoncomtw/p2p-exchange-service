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

func AppCreatePaymentMethodHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreatePaymentMethodRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		uidRaw, _ := r.Context().Value("sub").(json.Number)
		uid, _ := uidRaw.Int64()

		l := logic.NewAppCreatePaymentMethodLogic(r.Context(), svcCtx)
		resp, err := l.Create(uid, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}

func AppListPaymentMethodsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uidRaw, _ := r.Context().Value("sub").(json.Number)
		uid, _ := uidRaw.Int64()

		l := logic.NewAppListPaymentMethodsLogic(r.Context(), svcCtx)
		resp, err := l.List(uid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}

func AppDeletePaymentMethodHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID int64 `path:"id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		uidRaw, _ := r.Context().Value("sub").(json.Number)
		uid, _ := uidRaw.Int64()

		l := logic.NewAppDeletePaymentMethodLogic(r.Context(), svcCtx)
		err := l.Delete(uid, req.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(map[string]bool{"ok": true}))
		}
	}
}
