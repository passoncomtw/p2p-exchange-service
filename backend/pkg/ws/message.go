package ws

import "encoding/json"

const (
	// SubjectNotifyAdmin 是發布到 NATS P2P_NOTIFY stream 的 admin 廣播 subject。
	SubjectNotifyAdmin = "notify.admin.order"

	// SubjectNotifyBuyerPrefix / SubjectNotifySellerPrefix 用於 per-user 推送，
	// 完整 subject 為 prefix + strconv.FormatInt(userID, 10)。
	SubjectNotifyBuyerPrefix  = "notify.buyer."
	SubjectNotifySellerPrefix = "notify.seller."

	// WebSocket message type（前端 event.type 識別用）。
	EventOrderStatusChanged      = "order.status.changed"
	EventWalletBalanceChanged    = "wallet.balance.changed"
	EventFiatWithdrawalCreated   = "fiat.withdrawal.created"
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

// FiatWithdrawalCreatedData 是 fiat.withdrawal.created 事件的 payload（廣播給後台）。
type FiatWithdrawalCreatedData struct {
	WithdrawalID int64  `json:"withdrawal_id"`
	UserID       int64  `json:"user_id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
}

func NewMessage(eventType string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{Type: eventType, Data: data})
}
