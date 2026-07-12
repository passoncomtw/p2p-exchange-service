package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	apperrors "p2p-exchange/internal/errors"
)

const walletLockTTL = 5 * time.Second

// AcquireWalletLock 取得指定 user+currency 的分散式鎖。
// Key 格式：lock:wallet:{userID}:{currency}，TTL 5 秒。
// Redis 故障時降級（不阻塞業務），DB SELECT FOR UPDATE 仍提供保護。
func AcquireWalletLock(ctx context.Context, rdb redis.UniversalClient, userID int64, currency string) (func(), error) {
	key := fmt.Sprintf("lock:wallet:%d:%s", userID, currency)
	ok, err := rdb.SetNX(ctx, key, 1, walletLockTTL).Result()
	if err != nil {
		return func() {}, nil // degradation：Redis 故障時略過鎖
	}
	if !ok {
		return func() {}, apperrors.New(429, "wallet is being processed, please retry")
	}
	return func() {
		rdb.Del(context.Background(), key)
	}, nil
}

// AcquireWalletLocks 同時鎖定兩個錢包，按 userID 固定順序防止死鎖。
func AcquireWalletLocks(ctx context.Context, rdb redis.UniversalClient, idA, idB int64, currency string) (func(), error) {
	first, second := idA, idB
	if idA > idB {
		first, second = idB, idA
	}
	unlockFirst, err := AcquireWalletLock(ctx, rdb, first, currency)
	if err != nil {
		return func() {}, err
	}
	unlockSecond, err := AcquireWalletLock(ctx, rdb, second, currency)
	if err != nil {
		unlockFirst()
		return func() {}, err
	}
	return func() {
		unlockSecond()
		unlockFirst()
	}, nil
}
