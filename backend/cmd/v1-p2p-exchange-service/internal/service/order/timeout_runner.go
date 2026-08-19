package order_service

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/pkg/mq"
)

const orderTimeoutInterval = 60 * time.Second

// OrderTimeoutRunner 每 60 秒發布 order.timeout.check，觸發超時訂單取消流程。
type OrderTimeoutRunner struct {
	publisher mq.Publisher
}

func NewOrderTimeoutRunner(p mq.Publisher) *OrderTimeoutRunner {
	return &OrderTimeoutRunner{publisher: p}
}

func (r *OrderTimeoutRunner) Name() string { return "order-timeout" }

func (r *OrderTimeoutRunner) Run(ctx context.Context) {
	ticker := time.NewTicker(orderTimeoutInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.publisher.Publish(ctx, "order.timeout.check", nil); err != nil {
				logx.Errorf("order-timeout: publish error: %v", err)
			}
		}
	}
}
