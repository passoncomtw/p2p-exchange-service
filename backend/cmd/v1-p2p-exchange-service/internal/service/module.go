package service

import (
	authservice "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"
	backend_admin_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_admin"
	backend_auth_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_auth"
	listing_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/listing"
	order_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/order"
	payment_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/payment_method"

	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(authservice.New),
	fx.Provide(backend_auth_service.New),
	fx.Provide(backend_admin_service.New),
	fx.Provide(payment_service.New),
	fx.Provide(listing_service.New),
	fx.Provide(order_service.New),
)
