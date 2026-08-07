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
	ListingHandler       *handlers.ListingHandler
	OrderHandler         *handlers.OrderHandler
	BackendHandler       *handlers.BackendHandler
	V1Handler            *handlers.V1Handler
	WSHandler            *handlers.WSHandler
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
			{
				Method:  http.MethodPost,
				Path:    "/app/listings",
				Handler: p.ListingHandler.Create,
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/listings",
				Handler: p.ListingHandler.List,
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/listings/mine",
				Handler: p.ListingHandler.ListMine,
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/listings/:id",
				Handler: p.ListingHandler.Get,
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/listings/:id/cancel",
				Handler: p.ListingHandler.Cancel,
			},
			{
				Method:  http.MethodPost,
				Path:    "/app/orders",
				Handler: p.OrderHandler.Create,
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/orders",
				Handler: p.OrderHandler.List,
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/orders/:id",
				Handler: p.OrderHandler.Get,
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/pay",
				Handler: p.OrderHandler.Pay,
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/confirm",
				Handler: p.OrderHandler.Confirm,
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/cancel",
				Handler: p.OrderHandler.Cancel,
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/dispute",
				Handler: p.OrderHandler.Dispute,
			},
		},
		rest.WithJwt(p.Config.App.AccessSecret),
	)

	// backend public routes
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/backend/auth/login",
		Handler: p.BackendHandler.Login,
	})

	// backend admin routes (JWT: Backend.AccessSecret)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/backend/dashboard",
				Handler: p.BackendHandler.Dashboard,
			},
			{
				Method:  http.MethodGet,
				Path:    "/backend/listings",
				Handler: p.BackendHandler.ListListings,
			},
			{
				Method:  http.MethodGet,
				Path:    "/backend/orders",
				Handler: p.BackendHandler.ListOrders,
			},
			{
				Method:  http.MethodPut,
				Path:    "/backend/orders/:id/resolve",
				Handler: p.BackendHandler.ResolveOrder,
			},
		},
		rest.WithJwt(p.Config.Backend.AccessSecret),
	)

	// v1 legacy public routes
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/v1/orders",
			Handler: p.V1Handler.CreateOrder,
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/orders/mine",
			Handler: p.V1Handler.ListMyOrders,
		},
		{
			Method:  http.MethodPost,
			Path:    "/v1/orders/:id/cancel",
			Handler: p.V1Handler.CancelOrder,
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/admin/orders",
			Handler: p.V1Handler.AdminListOrders,
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/admin/orders/:id",
			Handler: p.V1Handler.AdminGetOrder,
		},
		{
			Method:  http.MethodPost,
			Path:    "/v1/admin/orders/:id/complete",
			Handler: p.V1Handler.AdminCompleteOrder,
		},
	})

	// WebSocket routes
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws/app",
		Handler: p.WSHandler.AppWS,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws/backend",
		Handler: p.WSHandler.BackendWS,
	})

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
