package job

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/infra/mq"
	pkgws "p2p-exchange/pkg/ws"
)

// WsEventDeps 是 WS 事件 consumer 的依賴。
type WsEventDeps struct {
	Hub *pkgws.Hub
}

// StartWsEventConsumer 訂閱 P2P_NOTIFY stream 的 notify.admin.* subject，
// 將訂單狀態事件廣播至所有後台 WebSocket 連線。
func StartWsEventConsumer(mqClient *mq.Client, deps WsEventDeps) {
	if mqClient == nil || deps.Hub == nil {
		return
	}
	if err := mqClient.Subscribe("notify.admin.*", func(_ context.Context, data []byte) error {
		return handleAdminNotify(data, deps.Hub)
	}); err != nil {
		logx.Errorf("ws event consumer: subscribe notify.admin.* error: %v", err)
	}
}

func handleAdminNotify(data []byte, hub *pkgws.Hub) error {
	var payload pkgws.OrderStatusChangedData
	if err := json.Unmarshal(data, &payload); err != nil {
		logx.Errorf("ws event: unmarshal admin notify error: %v", err)
		return nil
	}
	msg, err := pkgws.NewMessage(pkgws.EventOrderStatusChanged, payload)
	if err != nil {
		logx.Errorf("ws event: marshal message error: %v", err)
		return nil
	}
	hub.BroadcastToBackend(msg)
	return nil
}
