package service

import (
	authservice "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"
	listing_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/listing"
	payment_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/payment_method"

	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(authservice.New),
	fx.Provide(payment_service.New),
	fx.Provide(listing_service.New),
)
