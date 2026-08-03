package repository

import (
	escrowrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/escrow_record"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	orderstatuslogrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order_status_log"
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
	fx.Provide(orderrepo.New),
	fx.Provide(escrowrepo.New),
	fx.Provide(orderstatuslogrepo.New),
)
