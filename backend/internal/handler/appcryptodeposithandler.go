package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"p2p-exchange/internal/logic"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

func AppGetCryptoDepositInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAppGetCryptoDepositInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetDepositInfo(ctxUID(r))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func AppListCryptoDepositsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AppListCryptoDepositsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewAppListCryptoDepositsLogic(r.Context(), svcCtx)
		resp, err := l.List(ctxUID(r), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
