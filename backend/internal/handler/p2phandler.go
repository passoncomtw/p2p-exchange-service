// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"p2p-exchange/internal/logic"
	"p2p-exchange/internal/response"
	"p2p-exchange/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func P2pHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewP2pLogic(r.Context(), svcCtx)
		resp, err := l.P2p(nil)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}
