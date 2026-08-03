package service

import (
	authservice "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"

	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(authservice.New),
)
