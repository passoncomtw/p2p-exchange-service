package v1_interface

type V1Order struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Asset         string  `json:"asset"`
	Fiat          string  `json:"fiat"`
	Price         float64 `json:"price"`
	Quantity      float64 `json:"quantity"`
	TotalAmount   float64 `json:"totalAmount"`
	PaymentMethod string  `json:"paymentMethod"`
	Status        string  `json:"status"`
	CreatedBy     string  `json:"createdBy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type V1OrderListResponse struct {
	List []V1Order `json:"list"`
}
