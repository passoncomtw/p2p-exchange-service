package svc

import (
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	nats "github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/model"
)

type ServiceContext struct {
	Config         config.Config
	DB             sqlx.SqlConn
	Redis          redis.UniversalClient
	Js             nats.JetStreamContext
	AppUser        *model.AppUserModel
	BackendUser    *model.BackendUserModel
	PaymentMethod  *model.PaymentMethodModel
	Listing        *model.ListingModel
	Order          *model.OrderModel
	EscrowRecord   *model.EscrowRecordModel
	OrderStatusLog *model.OrderStatusLogModel
	V1Order        *model.V1OrderModel
	Wallet         *model.WalletModel
	WalletLedger   *model.WalletLedgerModel
	Currency       *model.CurrencyModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("pgx", c.Database.DSN)
	rdb := newRedisClient(c.Redis)
	js := newJetStream(c.Nats)
	return &ServiceContext{
		Config:         c,
		DB:             conn,
		Redis:          rdb,
		Js:             js,
		AppUser:        model.NewAppUserModel(conn),
		BackendUser:    model.NewBackendUserModel(conn),
		PaymentMethod:  model.NewPaymentMethodModel(conn),
		Listing:        model.NewListingModel(conn),
		Order:          model.NewOrderModel(conn),
		EscrowRecord:   model.NewEscrowRecordModel(conn),
		OrderStatusLog: model.NewOrderStatusLogModel(conn),
		V1Order:        model.NewV1OrderModel(conn),
		Wallet:         model.NewWalletModel(conn, rdb),
		WalletLedger:   model.NewWalletLedgerModel(conn),
		Currency:       model.NewCurrencyModel(conn),
	}
}

func newRedisClient(c config.RedisConf) redis.UniversalClient {
	if c.Addr == "" {
		return nil
	}
	addrs := splitTrimmed(c.Addr)
	poolSize := c.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    addrs,
		Password: c.Password,
		PoolSize: poolSize,
	})
	logx.Infof("redis connected: %s", c.Addr)
	return rdb
}

func newJetStream(c config.NatsConf) nats.JetStreamContext {
	if c.URL == "" {
		return nil
	}
	var opts []nats.Option
	if c.CredsPath != "" {
		opts = append(opts, nats.UserCredentials(c.CredsPath))
	}
	nc, err := nats.Connect(c.URL, opts...)
	if err != nil {
		logx.Errorf("nats connect %s error: %v", c.URL, err)
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		logx.Errorf("nats jetstream error: %v", err)
		return nil
	}
	if err := ensureStream(js, c); err != nil {
		logx.Errorf("nats ensure stream error: %v", err)
	}
	logx.Infof("nats jetstream connected: %s stream=%s", c.URL, c.StreamName)
	return js
}

func ensureStream(js nats.JetStreamContext, c config.NatsConf) error {
	_, err := js.StreamInfo(c.StreamName)
	if err == nil {
		return nil
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name: c.StreamName,
		Subjects: []string{
			"order.timeout.check",
			"order.status.changed",
			"wallet.balance.changed",
			"notification.push",
		},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	return err
}

func splitTrimmed(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}
	return result
}
