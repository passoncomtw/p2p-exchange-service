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

// New 依 config.RedisConf.Mode 建立對應的 Redis 連線：
//   - single   (預設): 單節點，Addr 為一個地址
//   - sentinel: 哨兵 HA，Addr 為 sentinel 地址列表，MasterName 為主節點名稱
//   - cluster  : 叢集，Addr 為多個節點地址
func New(c config.RedisConf) *Client {
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

	default: // single
		rdb = redis.NewClient(&redis.Options{
			Addr:     addrs[0],
			Password: c.Password,
			PoolSize: poolSize,
		})
		logx.Infof("redis single connected: %s", addrs[0])
	}

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
