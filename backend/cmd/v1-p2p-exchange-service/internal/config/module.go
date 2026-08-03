package config

import (
	postgresql "p2p-exchange/pkg/database_module/postgresql"
	redis "p2p-exchange/pkg/database_module/redis"
	"p2p-exchange/pkg/mq"

	"go.uber.org/fx"
)

var Module = fx.Module("config",
	fx.Provide(NewConfig),
	fx.Provide(func(cfg *Config) postgresql.Config {
		return postgresql.Config{DSN: cfg.Database.DSN}
	}),
	fx.Provide(func(cfg *Config) redis.Config {
		return redis.Config{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			PoolSize: cfg.Redis.PoolSize,
		}
	}),
	fx.Provide(func(cfg *Config) mq.Config {
		return mq.Config{
			URL:          cfg.Nats.URL,
			CredsPath:    cfg.Nats.CredsPath,
			User:         cfg.Nats.User,
			Password:     cfg.Nats.Password,
			StreamName:   cfg.Nats.StreamName,
			ConsumerName: cfg.Nats.ConsumerName,
		}
	}),
)
