package notification

import "context"

// INotificationSender 是各通知通道的共同介面。
// 目前實作：AppPushSender（Expo push）
// 未來擴充：WebSocketSender（Web 即時通知）
type INotificationSender interface {
	Send(ctx context.Context, msg Message) error
}

// Notifier 持有所有已註冊的 sender，Send 時 fan-out 給全部。
type Notifier struct {
	senders []INotificationSender
}

func New(senders ...INotificationSender) *Notifier {
	return &Notifier{senders: senders}
}

// Send 非同步地將訊息發送給所有已註冊的 sender，錯誤只 log 不阻斷。
func (n *Notifier) Send(ctx context.Context, msg Message) {
	for _, s := range n.senders {
		s := s
		go func() {
			if err := s.Send(ctx, msg); err != nil {
				// 各 sender 自行處理 error log，這裡不重複
				_ = err
			}
		}()
	}
}
