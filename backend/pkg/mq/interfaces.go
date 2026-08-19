package mq

import "context"

// MsgHandler 是 NATS consumer 的訊息處理函式。
// 回傳 nil → Ack；回傳 error → Nak（JetStream 自動重試）。
type MsgHandler func(ctx context.Context, data []byte) error

// Publisher 定義非同步/同步發布訊息到 NATS JetStream 的能力。
type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
	PublishAsync(subject string, data []byte)
}

// Subscriber 定義訂閱 NATS JetStream subject 的能力。
type Subscriber interface {
	Subscribe(subject string, handler MsgHandler) error
}
