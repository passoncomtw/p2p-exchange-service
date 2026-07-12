package rdb

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}
	return val, true, nil
}

func (c *Client) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, val, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// GetOrSet 先查快取，未命中時執行 fn 並回寫。fn 失敗時不寫快取，直接回傳錯誤。
func (c *Client) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	val, found, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if found {
		return val, nil
	}
	val, err = fn()
	if err != nil {
		return "", err
	}
	_ = c.Set(ctx, key, val, ttl)
	return val, nil
}
