package repository

import (
	paymentrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/payment_method"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"

	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(userrepo.New),
	fx.Provide(paymentrepo.New),
)
