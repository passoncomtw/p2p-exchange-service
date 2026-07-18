package notification

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/infra/mq"
)

const natsSubject = "notification.push"

// AppPushSender 將訊息發布到 NATS，由 push_notification_job consumer 負責實際送達 Expo。
type AppPushSender struct {
	mq *mq.Client
}

func NewAppPushSender(mqClient *mq.Client) *AppPushSender {
	return &AppPushSender{mq: mqClient}
}

func (s *AppPushSender) Send(_ context.Context, msg Message) error {
	if s.mq == nil {
		return nil
	}

	if msg.Channel == "" {
		msg.Channel = ChannelOrders
	}
	if msg.Priority == "" {
		msg.Priority = PriorityHigh
	}

	b, err := json.Marshal(msg)
	if err != nil {
		logx.Errorf("notification: marshal app push payload error: %v", err)
		return err
	}

	s.mq.PublishAsync(natsSubject, b)
	return nil
}
