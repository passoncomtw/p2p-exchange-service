package notification

const (
	ChannelOrders = "orders"
	ChannelSystem = "system"

	PriorityDefault = "default"
	PriorityNormal  = "normal"
	PriorityHigh    = "high"
)

type Message struct {
	RecipientID int64  `json:"recipient_id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	OrderID     int64  `json:"order_id"`
	Channel     string `json:"channel,omitempty"`
	Priority    string `json:"priority,omitempty"`
}
