package handler

import (
	"net/http"

	"p2p-exchange/internal/logic"
	"p2p-exchange/internal/svc"
)

func WebhookECPayNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "0|parse error", http.StatusBadRequest)
			return
		}
		params := make(map[string]string, len(r.Form))
		for k, vs := range r.Form {
			if len(vs) > 0 {
				params[k] = vs[0]
			}
		}
		l := logic.NewWebhookECPayLogic(r.Context(), svcCtx)
		if err := l.HandleNotify(params); err != nil {
			http.Error(w, "0|"+err.Error(), http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("1|OK"))
	}
}
