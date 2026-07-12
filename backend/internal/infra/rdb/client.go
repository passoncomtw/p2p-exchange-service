package rdb

import (
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/config"
)

type Client struct {
	rdb redis.UniversalClient
}

func New(c config.RedisConf) *Client {
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
	return &Client{rdb: rdb}
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
