package job

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/infra/mq"
)

const orderTimeoutCheckInterval = 60 * time.Second

// StartScheduler 每 60 秒 publish order.timeout.check，由 consumer 處理超時訂單。
// NATS 不可用時自動略過。
func StartScheduler(ctx context.Context, mqClient *mq.Client) {
	if mqClient == nil {
		logx.Info("NATS not configured, scheduler disabled")
		return
	}
	mqClient.StartSchedule(ctx, "order.timeout.check", orderTimeoutCheckInterval)
}
