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

// Subscribe 建立 Durable Consumer 訂閱 subject，僅處理訂閱後的新訊息（DeliverNew）。
// 若既有 consumer 設定與目前不符（例如舊版使用 DeliverAll），自動刪除並重建。
// handler 回 nil → Ack；回 error → Nak（JetStream 自動重試）。
// 內建 panic recover，防止 handler panic 中斷整個 consumer。
func (c *Consumer) Subscribe(subject string, handler MsgHandler) error {
	if c.client == nil || c.client.js == nil {
		return nil
	}
	consumerName := buildConsumerName(c.client.consumerName, subject)
	wrap := c.buildHandler(subject, handler)

	_, err := c.client.js.Subscribe(subject, wrap,
		nats.Durable(consumerName),
		nats.ManualAck(),
		nats.DeliverNew(),
	)
	if err != nil && strings.Contains(err.Error(), "configuration comparisons mismatch") {
		logx.Infof("subscriber %s: consumer config changed, recreating with DeliverNew", subject)
		if delErr := c.client.js.DeleteConsumer(c.client.streamName, consumerName); delErr != nil {
			return fmt.Errorf("delete stale consumer %s: %w", consumerName, delErr)
		}
		_, err = c.client.js.Subscribe(subject, wrap,
			nats.Durable(consumerName),
			nats.ManualAck(),
			nats.DeliverNew(),
		)
	}
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	return nil
}

func (c *Consumer) buildHandler(subject string, handler MsgHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
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
	}
}

// buildConsumerName 將 subject 轉為合法的 NATS durable consumer name。
// NATS consumer name 不允許 . 或 *，統一替換。
func buildConsumerName(base, subject string) string {
	sanitized := strings.NewReplacer(".", "-", "*", "wildcard").Replace(subject)
	return base + "-" + sanitized
}
