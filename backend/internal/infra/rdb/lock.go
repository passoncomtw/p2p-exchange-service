package rdb

import (
	"context"
	"sort"
	"time"

	apperrors "p2p-exchange/internal/errors"
)

// AcquireLock 取得指定 key 的分散式鎖，TTL 後自動釋放。
// Redis 故障時降級（unlock 為 no-op），DB SELECT FOR UPDATE 仍提供保護。
// 鎖被佔用時回 429，由呼叫方決定錯誤處理。
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (unlock func(), err error) {
	ok, err := c.rdb.SetNX(ctx, key, 1, ttl).Result()
	if err != nil {
		return func() {}, nil
	}
	if !ok {
		return func() {}, apperrors.New(429, "resource is being processed, please retry")
	}
	return func() {
		c.rdb.Del(context.Background(), key)
	}, nil
}

// AcquireLocks 同時鎖定多個 key，按字母排序保證固定順序防止死鎖。
// 任一 key 失敗時釋放已取得的鎖並回傳錯誤。
func (c *Client) AcquireLocks(ctx context.Context, keys []string, ttl time.Duration) (unlock func(), err error) {
	sorted := make([]string, len(keys))
	copy(sorted, keys)
	sort.Strings(sorted)

	unlocks := make([]func(), 0, len(sorted))
	for _, key := range sorted {
		ul, err := c.AcquireLock(ctx, key, ttl)
		if err != nil {
			for i := len(unlocks) - 1; i >= 0; i-- {
				unlocks[i]()
			}
			return func() {}, err
		}
		unlocks = append(unlocks, ul)
	}
	return func() {
		for i := len(unlocks) - 1; i >= 0; i-- {
			unlocks[i]()
		}
	}, nil
}
