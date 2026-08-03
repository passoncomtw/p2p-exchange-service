package redis

import (
	"fmt"
	"p2p-exchange/internal/infra/rdb"

	"go.uber.org/fx"
)

var Module = fx.Module("redis",
	fx.Provide(NewRedis),
)

type Config struct {
	Mode       string
	Addr       string
	Password   string
	MasterName string
	PoolSize   int
}

func NewRedis(cfg Config) *rdb.Client {
	redisClient := rdb.NewFromConfig(rdb.Config{
		Mode:       cfg.Mode,
		Addr:       cfg.Addr,
		Password:   cfg.Password,
		MasterName: cfg.MasterName,
		PoolSize:   cfg.PoolSize,
	})
	fmt.Println("redis is connected")
	return redisClient
}
