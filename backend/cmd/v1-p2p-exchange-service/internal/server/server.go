package server

import (
	"context"
	"fmt"
	"net/http"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/server/handlers"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/swagger"
	"p2p-exchange/internal/response"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Config               *config.Config
	LoginHandler         *handlers.LoginHandler
	ProfileHandler       *handlers.ProfileHandler
	PaymentMethodHandler *handlers.PaymentMethodHandler
	LC                   fx.Lifecycle
}

func NewServer(p Params) *rest.Server {
	server := rest.MustNewServer(p.Config.RestConf,
		rest.WithCors("*"),
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, _ error) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "Unauthorized"))
		}),
	)

	// public routes
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/version",
		Handler: handlers.VersionHandler,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/app/auth/login",
		Handler: p.LoginHandler.Handle,
	})

	// app private routes (JWT: App.AccessSecret)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/app/profile",
				Handler: p.ProfileHandler.Handle,
			},
			{
				Method:  http.MethodPost,
				Path:    "/app/payment-methods",
				Handler: p.PaymentMethodHandler.Create,
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/payment-methods",
				Handler: p.PaymentMethodHandler.List,
			},
			{
				Method:  http.MethodDelete,
				Path:    "/app/payment-methods/:id",
				Handler: p.PaymentMethodHandler.Delete,
			},
		},
		rest.WithJwt(p.Config.App.AccessSecret),
	)

	if p.Config.RestConf.Mode != "prod" {
		swagger.RegisterRoutes(server)
		fmt.Printf("Swagger UI: http://localhost:%d/swagger\n", p.Config.RestConf.Port)
	}

	p.LC.Append(fx.Hook{
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
