package v1_interface

type V1CreateOrderRequest struct {
	Type          string  `json:"type"`
	Asset         string  `json:"asset"`
	Fiat          string  `json:"fiat"`
	Price         float64 `json:"price"`
	Quantity      float64 `json:"quantity"`
	PaymentMethod string  `json:"paymentMethod"`
}

type V1OrderIDRequest struct {
	ID int64 `path:"id"`
}

type V1AdminListOrdersRequest struct {
	Status string `form:"status,optional"`
}
