package app_interface

type CreateOrderResponse struct {
	ID      int64  `json:"id"`
	OrderNo string `json:"orderNo"`
}

type OrderItem struct {
	ID                int64    `json:"id"`
	OrderNo           string   `json:"orderNo"`
	ListingID         int64    `json:"listingId"`
	ListingType       string   `json:"listingType"`
	SellerID          int64    `json:"sellerId"`
	BuyerID           int64    `json:"buyerId"`
	CryptoCurrency    string   `json:"cryptoCurrency"`
	FiatCurrency      string   `json:"fiatCurrency"`
	CryptoAmount      float64  `json:"cryptoAmount"`
	Price             float64  `json:"price"`
	FiatAmount        float64  `json:"fiatAmount"`
	PlatformFeeBase   float64  `json:"platformFeeBase"`
	PlatformFeeAmount float64  `json:"platformFeeAmount"`
	PaymentFeeBase    float64  `json:"paymentFeeBase"`
	PaymentFeeAmount  float64  `json:"paymentFeeAmount"`
	TotalFee          float64  `json:"totalFee"`
	TotalAmount       float64  `json:"totalAmount"`
	PaymentMethodID   int64    `json:"paymentMethodId"`
	Status            string   `json:"status"`
	PaymentDeadline   string   `json:"paymentDeadline"`
	PaidAt            *string  `json:"paidAt,omitempty"`
	ConfirmedAt       *string  `json:"confirmedAt,omitempty"`
	CompletedAt       *string  `json:"completedAt,omitempty"`
	CancelledAt       *string  `json:"cancelledAt,omitempty"`
	CancelReason      *string  `json:"cancelReason,omitempty"`
	CreatedAt         string   `json:"createdAt"`
}

type ListOrdersResponse struct {
	List []OrderItem `json:"list"`
}
