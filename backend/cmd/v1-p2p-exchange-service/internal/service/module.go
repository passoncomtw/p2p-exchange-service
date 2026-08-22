package service

import (
	authservice "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/auth"
	backend_admin_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_admin"
	backend_auth_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/backend_auth"
	cryptodeposit_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/crypto_deposit"
	fiatdeposit_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/fiat_deposit"
	fiatwithdraw_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/fiat_withdraw"
	listing_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/listing"
	notification_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/notification"
	order_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/order"
	payment_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/payment_method"
	v1_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/v1"
	wallet_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/wallet"
	"p2p-exchange/pkg/schedule"

	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(authservice.New),
	fx.Provide(backend_auth_service.New),
	fx.Provide(backend_admin_service.New),
	fx.Provide(payment_service.New),
	fx.Provide(listing_service.New),
	fx.Provide(order_service.New),
	fx.Provide(v1_service.New),
	fx.Provide(wallet_service.New),
	fx.Provide(fiatdeposit_service.New),
	fx.Provide(fiatwithdraw_service.New),
	fx.Provide(cryptodeposit_service.New),
	// 定時任務：注冊進 schedule_runner group，由 Scheduler 統一啟停
	fx.Provide(fx.Annotate(
		order_service.NewOrderTimeoutRunner,
		fx.As(new(schedule.Runner)),
		fx.ResultTags(`group:"schedule_runner"`),
	)),
	fx.Provide(fx.Annotate(
		cryptodeposit_service.NewTronScannerRunner,
		fx.As(new(schedule.Runner)),
		fx.ResultTags(`group:"schedule_runner"`),
	)),
	// NATS Consumer：fx.Invoke 在啟動時訂閱 subject
	fx.Invoke(order_service.RegisterExpiredOrderConsumer),
	fx.Invoke(notification_service.RegisterWsEventConsumer),
	fx.Invoke(notification_service.RegisterPushNotificationConsumer),
)
