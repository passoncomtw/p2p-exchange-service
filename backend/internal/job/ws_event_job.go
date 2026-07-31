package job

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/infra/mq"
	pkgws "p2p-exchange/pkg/ws"
)

// WsEventDeps 是 WS 事件 consumer 的依賴。
type WsEventDeps struct {
	Hub *pkgws.Hub
}

// StartWsEventConsumer 訂閱 P2P_NOTIFY stream 的三類 subject：
//   - notify.admin.*  → 廣播至所有後台 WebSocket 連線
//   - notify.buyer.*  → 推送至對應 buyer 的 App 連線
//   - notify.seller.* → 推送至對應 seller 的 App 連線
func StartWsEventConsumer(mqClient *mq.Client, deps WsEventDeps) {
	if mqClient == nil || deps.Hub == nil {
		return
	}
	if err := mqClient.Subscribe("notify.admin.*", func(_ context.Context, data []byte) error {
		return handleAdminNotify(data, deps.Hub)
	}); err != nil {
		logx.Errorf("ws event consumer: subscribe notify.admin.* error: %v", err)
	}
	if err := mqClient.Subscribe("notify.buyer.*", func(_ context.Context, data []byte) error {
		return handleUserNotify(data, deps.Hub, pkgws.SubjectNotifyBuyerPrefix)
	}); err != nil {
		logx.Errorf("ws event consumer: subscribe notify.buyer.* error: %v", err)
	}
	if err := mqClient.Subscribe("notify.seller.*", func(_ context.Context, data []byte) error {
		return handleUserNotify(data, deps.Hub, pkgws.SubjectNotifySellerPrefix)
	}); err != nil {
		logx.Errorf("ws event consumer: subscribe notify.seller.* error: %v", err)
	}
}

// handleUserNotify 解析 buyer/seller 事件並推送至指定 user 的 App 連線。
// subjectPrefix 用於從 payload 欄位判斷推送對象（buyer 或 seller）。
func handleUserNotify(data []byte, hub *pkgws.Hub, subjectPrefix string) error {
	var payload pkgws.OrderStatusChangedData
	if err := json.Unmarshal(data, &payload); err != nil {
		logx.Errorf("ws event: unmarshal user notify error: %v", err)
		return nil
	}
	msg, err := pkgws.NewMessage(pkgws.EventOrderStatusChanged, payload)
	if err != nil {
		logx.Errorf("ws event: marshal message error: %v", err)
		return nil
	}
	var userID int64
	switch {
	case strings.HasPrefix(subjectPrefix, pkgws.SubjectNotifyBuyerPrefix):
		userID = payload.BuyerID
	case strings.HasPrefix(subjectPrefix, pkgws.SubjectNotifySellerPrefix):
		userID = payload.SellerID
	default:
		logx.Errorf("ws event: unknown subject prefix: %s", subjectPrefix)
		return nil
	}
	if userID <= 0 {
		logx.Errorf("ws event: invalid userID %d for prefix %s", userID, subjectPrefix)
		return nil
	}
	logx.Infof("ws event: sending order.status.changed to user %s", strconv.FormatInt(userID, 10))
	hub.SendToUser(userID, msg)
	return nil
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
