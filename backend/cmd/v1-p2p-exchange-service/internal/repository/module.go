package repository

import (
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	paymentrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/payment_method"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"

	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(userrepo.New),
	fx.Provide(paymentrepo.New),
	fx.Provide(listingrepo.New),
	fx.Provide(walletrepo.New),
)
