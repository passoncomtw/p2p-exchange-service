package server

import (
	"context"
	"net/http"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/server/handlers"
	"p2p-exchange/internal/response"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.uber.org/fx"
)

func NewServer(config *config.Config, lc fx.Lifecycle) *rest.Server {
	server := rest.MustNewServer(config.RestConf,
		rest.WithCors("*"),
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, _ error) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "Unauthorized"))
		}),
	)

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/version",
		Handler: handlers.VersionHandler,
	})

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go server.Start()
			return nil
		},
		OnStop: func(_ context.Context) error {
			server.Stop()
			return nil
		},
	})

	return server
}
