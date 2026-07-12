package handler

import (
	"net/http"

	"p2p-exchange/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// ── v1 掛單（免登入，沿用 listings；createdBy 固定 demo_user） ──────────────
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/v1/orders",
			Handler: V1CreateOrderHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/orders/mine",
			Handler: V1ListMyOrdersHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/v1/orders/:id/cancel",
			Handler: V1CancelOrderHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/admin/orders",
			Handler: V1AdminListOrdersHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/v1/admin/orders/:id",
			Handler: V1AdminGetOrderHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/v1/admin/orders/:id/complete",
			Handler: V1AdminCompleteOrderHandler(serverCtx),
		},
	})

	// ── app public ──────────────────────────────────────────────────────────
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/app/auth/login",
			Handler: AppLoginHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/app/auth/register",
			Handler: AppRegisterHandler(serverCtx),
		},
	})

	// ── app private (JWT: App.AccessSecret) ─────────────────────────────────
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/app/profile",
				Handler: AppProfileHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/app/payment-methods",
				Handler: AppCreatePaymentMethodHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/payment-methods",
				Handler: AppListPaymentMethodsHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/app/payment-methods/:id",
				Handler: AppDeletePaymentMethodHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/app/listings",
				Handler: AppCreateListingHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/listings",
				Handler: AppListListingsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/listings/mine",
				Handler: AppMyListingsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/listings/:id",
				Handler: AppGetListingHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/listings/:id/cancel",
				Handler: AppCancelListingHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/app/orders",
				Handler: AppCreateOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/orders",
				Handler: AppListOrdersHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/orders/:id",
				Handler: AppGetOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/pay",
				Handler: AppPayOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/confirm",
				Handler: AppConfirmOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/cancel",
				Handler: AppCancelOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/app/orders/:id/dispute",
				Handler: AppDisputeOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/wallets",
				Handler: AppListWalletsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/app/wallets/:currency/ledgers",
				Handler: AppListWalletLedgersHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.App.AccessSecret),
	)

	// ── backend public ──────────────────────────────────────────────────────
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/backend/auth/login",
			Handler: BackendLoginHandler(serverCtx),
		},
	})

	// ── backend private (JWT: Backend.AccessSecret) ──────────────────────────
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/backend/members",
				Handler: BackendListMembersHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/backend/dashboard",
				Handler: BackendDashboardHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/backend/listings",
				Handler: BackendListListingsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/backend/orders",
				Handler: BackendListOrdersHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/backend/orders/:id/resolve",
				Handler: BackendResolveOrderHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Backend.AccessSecret),
	)
}
