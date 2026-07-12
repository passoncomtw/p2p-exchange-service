package mq

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartSchedule 在背景 goroutine 中依照 interval 週期性 publish 到 subject。
// ctx.Done() 時停止。
func (c *Client) StartSchedule(ctx context.Context, subject string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.Publish(ctx, subject, nil); err != nil {
					logx.Errorf("scheduler: publish %s error: %v", subject, err)
				}
			}
		}
	}()
}
