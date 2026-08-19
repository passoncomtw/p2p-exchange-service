package notification_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/fx"

	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	"p2p-exchange/pkg/mq"
	"p2p-exchange/pkg/notification"
)

const expoPushAPIURL = "https://exp.host/--/api/v2/push/send"

type expoPushPayload struct {
	To        string                 `json:"to"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data,omitempty"`
	ChannelID string                 `json:"channelId,omitempty"`
	Priority  string                 `json:"priority,omitempty"`
}

type PushNotificationDeps struct {
	fx.In

	Subscriber mq.Subscriber
	UserRepo   userrepo.UserRepository
}

// RegisterPushNotificationConsumer 訂閱 notification.push，呼叫 Expo API 發送手機推播。
func RegisterPushNotificationConsumer(deps PushNotificationDeps) {
	if err := deps.Subscriber.Subscribe("notification.push", func(ctx context.Context, data []byte) error {
		return handlePush(ctx, data, deps.UserRepo)
	}); err != nil {
		logx.Errorf("push-notification: subscribe error: %v", err)
	}
}

func handlePush(ctx context.Context, data []byte, userRepo userrepo.UserRepository) error {
	var msg notification.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		logx.Errorf("push-notification: unmarshal error: %v", err)
		return nil
	}

	token, err := userRepo.GetPushToken(ctx, msg.RecipientID)
	if err != nil || token == "" {
		return nil
	}

	channel := msg.Channel
	if channel == "" {
		channel = notification.ChannelOrders
	}
	priority := msg.Priority
	if priority == "" {
		priority = notification.PriorityHigh
	}

	payload := []expoPushPayload{{
		To:        token,
		Title:     msg.Title,
		Body:      msg.Body,
		Data:      map[string]interface{}{"orderId": msg.OrderID},
		ChannelID: channel,
		Priority:  priority,
	}}

	b, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(expoPushAPIURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("expo push http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logx.Errorf("push-notification: expo API error: status=%d recipientId=%d", resp.StatusCode, msg.RecipientID)
	}
	return nil
}
