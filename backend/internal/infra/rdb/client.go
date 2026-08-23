package rdb

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type Client struct {
	rdb redis.UniversalClient
}

type Config struct {
	Mode       string
	Addr       string
	Password   string
	MasterName string
	PoolSize   int
}

func NewFromConfig(c Config) *Client {
	if c.Addr == "" {
		return nil
	}
	poolSize := c.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	var rdb redis.UniversalClient
	addrs := splitTrimmed(c.Addr)

	switch strings.ToLower(c.Mode) {
	case "sentinel":
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    c.MasterName,
			SentinelAddrs: addrs,
			Password:      c.Password,
			PoolSize:      poolSize,
		})
		logx.Infof("redis sentinel connected: master=%s sentinels=%s", c.MasterName, c.Addr)

	case "cluster":
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: c.Password,
			PoolSize: poolSize,
		})
		logx.Infof("redis cluster connected: %s", c.Addr)

	default:
		rdb = redis.NewClient(&redis.Options{
			Addr:     addrs[0],
			Password: c.Password,
			PoolSize: poolSize,
		})
		logx.Infof("redis single connected: %s", addrs[0])
	}

	return &Client{rdb: rdb}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
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
