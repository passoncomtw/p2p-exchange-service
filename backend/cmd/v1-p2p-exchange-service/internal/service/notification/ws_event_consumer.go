package notification_service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/fx"

	"p2p-exchange/pkg/mq"
	pkgws "p2p-exchange/pkg/ws"
)

type WsEventDeps struct {
	fx.In

	Subscriber mq.Subscriber
	Hub        *pkgws.Hub
}

// RegisterWsEventConsumer 訂閱 P2P_NOTIFY stream，將事件推送至對應的 WebSocket 連線。
func RegisterWsEventConsumer(deps WsEventDeps) {
	subjects := []struct {
		subject string
		handler func([]byte) error
	}{
		{"notify.admin.*", func(data []byte) error { return pushToBackend(data, deps.Hub) }},
		{"notify.buyer.*", func(data []byte) error { return pushToUser(data, deps.Hub, "buyer") }},
		{"notify.seller.*", func(data []byte) error { return pushToUser(data, deps.Hub, "seller") }},
	}

	for _, s := range subjects {
		subject := s.subject
		handler := s.handler
		if err := deps.Subscriber.Subscribe(subject, func(_ context.Context, data []byte) error {
			return handler(data)
		}); err != nil {
			logx.Errorf("ws-event: subscribe %s error: %v", subject, err)
		}
	}
}

func pushToBackend(data []byte, hub *pkgws.Hub) error {
	var payload pkgws.OrderStatusChangedData
	if err := json.Unmarshal(data, &payload); err != nil {
		logx.Errorf("ws-event: unmarshal admin notify error: %v", err)
		return nil
	}
	msg, err := pkgws.NewMessage(pkgws.EventOrderStatusChanged, payload)
	if err != nil {
		return nil
	}
	hub.BroadcastToBackend(msg)
	return nil
}

func pushToUser(data []byte, hub *pkgws.Hub, role string) error {
	var payload pkgws.OrderStatusChangedData
	if err := json.Unmarshal(data, &payload); err != nil {
		logx.Errorf("ws-event: unmarshal %s notify error: %v", role, err)
		return nil
	}
	msg, err := pkgws.NewMessage(pkgws.EventOrderStatusChanged, payload)
	if err != nil {
		return nil
	}
	var userID int64
	switch strings.ToLower(role) {
	case "buyer":
		userID = payload.BuyerID
	case "seller":
		userID = payload.SellerID
	}
	if userID <= 0 {
		logx.Errorf("ws-event: invalid userID for role %s", role)
		return nil
	}
	logx.Infof("ws-event: sending to user %s", strconv.FormatInt(userID, 10))
	hub.SendToUser(userID, msg)
	return nil
}
