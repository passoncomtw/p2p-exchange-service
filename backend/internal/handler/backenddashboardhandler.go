package handler

import (
	"net/http"

	"p2p-exchange/internal/logic"
	"p2p-exchange/internal/response"
	"p2p-exchange/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func BackendDashboardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBackendDashboardLogic(r.Context(), svcCtx)
		resp, err := l.Dashboard()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, response.Success(resp))
		}
	}
}
