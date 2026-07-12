package job

import (
	"context"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
)

const orderTimeoutCheckInterval = 60 * time.Second

// StartScheduler 每 60 秒 publish order.timeout.check，由 consumer 處理超時訂單。
// NATS 不可用時自動略過（不影響主服務）。
func StartScheduler(ctx context.Context, js nats.JetStreamContext) {
	if js == nil {
		logx.Info("NATS not configured, scheduler disabled")
		return
	}
	go func() {
		ticker := time.NewTicker(orderTimeoutCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := js.Publish("order.timeout.check", nil); err != nil {
					logx.Errorf("scheduler: publish order.timeout.check error: %v", err)
				}
			}
		}
	}()
}
