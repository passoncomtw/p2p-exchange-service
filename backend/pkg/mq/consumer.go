package mq

import (
	"context"
	"fmt"
	"strings"

	nats "github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
)

// Consumer 實作 Subscriber interface，封裝 JetStream durable consumer 邏輯。
type Consumer struct {
	client *Client
}

func NewConsumer(c *Client) *Consumer {
	return &Consumer{client: c}
}

// Subscribe 建立 Durable Consumer 訂閱 subject。
// handler 回 nil → Ack；回 error → Nak（JetStream 自動重試）。
// 內建 panic recover，防止 handler panic 中斷整個 consumer。
func (c *Consumer) Subscribe(subject string, handler MsgHandler) error {
	if c.client == nil || c.client.js == nil {
		return nil
	}
	consumerName := buildConsumerName(c.client.consumerName, subject)
	_, err := c.client.js.Subscribe(subject, func(msg *nats.Msg) {
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
	}, nats.Durable(consumerName), nats.ManualAck())
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	return nil
}

// buildConsumerName 將 subject 轉為合法的 NATS durable consumer name。
// NATS consumer name 不允許 . 或 *，統一替換。
func buildConsumerName(base, subject string) string {
	sanitized := strings.NewReplacer(".", "-", "*", "wildcard").Replace(subject)
	return base + "-" + sanitized
}
