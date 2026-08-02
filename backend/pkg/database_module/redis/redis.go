package redis

import (
	"fmt"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/infra/rdb"

	"go.uber.org/fx"
)

var Module = fx.Module("redis",
	fx.Provide(NewRedis),
)

func NewRedis(config config.Config) *rdb.Client {
	redisClient := rdb.New(config.Redis)
	fmt.Println("redis is connected")
	return redisClient
}
