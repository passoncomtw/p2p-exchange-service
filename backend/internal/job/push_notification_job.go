package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/infra/mq"
	"p2p-exchange/internal/model"
)

const expoAPIURL = "https://exp.host/--/api/v2/push/send"

type PushNotificationMessage struct {
	RecipientID int64  `json:"recipient_id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	OrderID     int64  `json:"order_id"`
}

type expoPushPayload struct {
	To    string                 `json:"to"`
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

type PushNotificationDeps struct {
	AppUser *model.AppUserModel
}

func StartPushNotificationConsumer(mqClient *mq.Client, deps PushNotificationDeps) {
	if mqClient == nil {
		return
	}
	if err := mqClient.Subscribe("notification.push", func(ctx context.Context, data []byte) error {
		return handlePushNotification(ctx, data, deps)
	}); err != nil {
		logx.Errorf("subscribe notification.push error: %v", err)
	}
}

func handlePushNotification(ctx context.Context, data []byte, deps PushNotificationDeps) error {
	var msg PushNotificationMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logx.Errorf("push notification: unmarshal error: %v", err)
		return nil // bad message, don't retry
	}

	token, err := deps.AppUser.GetPushToken(ctx, msg.RecipientID)
	if err != nil || token == "" {
		return nil // user has no token, skip silently
	}

	payload := []expoPushPayload{{
		To:    token,
		Title: msg.Title,
		Body:  msg.Body,
		Data:  map[string]interface{}{"orderId": msg.OrderID},
	}}

	b, _ := json.Marshal(payload)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Post(expoAPIURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("expo push http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logx.Errorf("expo push API error: status=%d recipientId=%d", resp.StatusCode, msg.RecipientID)
	}
	return nil
}
