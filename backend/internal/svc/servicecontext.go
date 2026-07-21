package svc

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/infra/mq"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/internal/model"
	"p2p-exchange/pkg/notification"
	pkgws "p2p-exchange/pkg/ws"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	RDB             *rdb.Client
	MQ              *mq.Client
	Notifier        *notification.Notifier
	Hub             *pkgws.Hub
	AppUser         *model.AppUserModel
	BackendUser     *model.BackendUserModel
	PaymentMethod   *model.PaymentMethodModel
	Listing         *model.ListingModel
	Order           *model.OrderModel
	EscrowRecord    *model.EscrowRecordModel
	OrderStatusLog  *model.OrderStatusLogModel
	V1Order         *model.V1OrderModel
	Wallet          *model.WalletModel
	WalletLedger    *model.WalletLedgerModel
	Currency        *model.CurrencyModel
	CryptoDeposit   *model.CryptoDepositModel
	CryptoWithdraw  *model.CryptoWithdrawalModel
	FiatDeposit     *model.FiatDepositModel
	FiatWithdraw    *model.FiatWithdrawalModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("pgx", c.Database.DSN)
	redisClient := rdb.New(c.Redis)
	mqClient := mq.New(c.Nats)
	return &ServiceContext{
		Config:         c,
		DB:             conn,
		RDB:            redisClient,
		MQ:             mqClient,
		Notifier:       notification.New(notification.NewAppPushSender(mqClient)),
		Hub:            pkgws.NewHub(),
		AppUser:        model.NewAppUserModel(conn),
		BackendUser:    model.NewBackendUserModel(conn),
		PaymentMethod:  model.NewPaymentMethodModel(conn),
		Listing:        model.NewListingModel(conn),
		Order:          model.NewOrderModel(conn),
		EscrowRecord:   model.NewEscrowRecordModel(conn),
		OrderStatusLog: model.NewOrderStatusLogModel(conn),
		V1Order:        model.NewV1OrderModel(conn),
		Wallet:         model.NewWalletModel(conn, redisClient),
		WalletLedger:   model.NewWalletLedgerModel(conn),
		Currency:       model.NewCurrencyModel(conn),
		CryptoDeposit:  model.NewCryptoDepositModel(conn),
		CryptoWithdraw: model.NewCryptoWithdrawalModel(conn),
		FiatDeposit:    model.NewFiatDepositModel(conn),
		FiatWithdraw:   model.NewFiatWithdrawalModel(conn),
	}
}
