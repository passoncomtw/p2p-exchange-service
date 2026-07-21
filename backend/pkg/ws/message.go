package ws

import "encoding/json"

const (
	EventOrderStatusChanged  = "order.status.changed"
	EventWalletBalanceChanged = "wallet.balance.changed"
)

// Message 是 WebSocket 推送的統一訊息格式。
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// OrderStatusChangedData 是 order.status.changed 事件的 payload。
type OrderStatusChangedData struct {
	OrderID  int64  `json:"order_id"`
	Status   string `json:"status"`
	BuyerID  int64  `json:"buyer_id"`
	SellerID int64  `json:"seller_id"`
}

// WalletBalanceChangedData 是 wallet.balance.changed 事件的 payload。
type WalletBalanceChangedData struct {
	UserID   int64  `json:"user_id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
}

func NewMessage(eventType string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{Type: eventType, Data: data})
}
