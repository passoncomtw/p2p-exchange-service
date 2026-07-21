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

// StartWsEventConsumer 訂閱 NATS 事件並推送到 WebSocket 連線。
func StartWsEventConsumer(mqClient *mq.Client, deps WsEventDeps) {
	if mqClient == nil || deps.Hub == nil {
		return
	}
	subjects := []string{
		pkgws.EventOrderStatusChanged,
		pkgws.EventWalletBalanceChanged,
	}
	for _, subj := range subjects {
		subj := subj
		if err := mqClient.Subscribe(subj, func(_ context.Context, data []byte) error {
			return handleWsEvent(subj, data, deps.Hub)
		}); err != nil {
			logx.Errorf("ws event consumer: subscribe %s error: %v", subj, err)
		}
	}
}

func handleWsEvent(subject string, data []byte, hub *pkgws.Hub) error {
	msg, err := pkgws.NewMessage(subject, json.RawMessage(data))
	if err != nil {
		logx.Errorf("ws event: marshal message error: %v", err)
		return nil
	}

	switch subject {
	case pkgws.EventOrderStatusChanged:
		var payload pkgws.OrderStatusChangedData
		if err := json.Unmarshal(data, &payload); err != nil {
			logx.Errorf("ws event: unmarshal order.status.changed error: %v", err)
			return nil
		}
		hub.SendToUser(payload.BuyerID, msg)
		hub.SendToUser(payload.SellerID, msg)
		hub.BroadcastToBackend(msg)

	case pkgws.EventWalletBalanceChanged:
		var payload pkgws.WalletBalanceChangedData
		if err := json.Unmarshal(data, &payload); err != nil {
			logx.Errorf("ws event: unmarshal wallet.balance.changed error: %v", err)
			return nil
		}
		hub.SendToUser(payload.UserID, msg)
	}
	return nil
}
