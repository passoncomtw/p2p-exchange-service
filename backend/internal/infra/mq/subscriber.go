package mq

import (
	"context"
	"fmt"

	nats "github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
)

type MsgHandler func(ctx context.Context, data []byte) error

// Subscribe 建立 Durable Consumer 訂閱 subject。
// handler 回 nil → Ack；回 error → Nak（JetStream 自動重試）。
// 內建 panic recover，防止 handler panic 打死 consumer。
func (c *Client) Subscribe(subject string, handler MsgHandler) error {
	_, err := c.js.Subscribe(subject, func(msg *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("subscriber panic on %s: %v", subject, r)
				_ = msg.Nak()
			}
		}()
		if err := handler(context.Background(), msg.Data); err != nil {
			logx.Errorf("subscriber %s handler error: %v", subject, err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	}, nats.Durable(c.consumerName), nats.ManualAck())
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	return nil
}
