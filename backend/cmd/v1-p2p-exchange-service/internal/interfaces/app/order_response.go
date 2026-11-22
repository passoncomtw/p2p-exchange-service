package app_interface

type CreateOrderResponse struct {
	ID      int64  `json:"id" example:"1"`
	OrderNo string `json:"orderNo" example:"P2P20261121123456"`
}

type OrderItem struct {
	ID                int64   `json:"id" example:"1"`
	OrderNo           string  `json:"orderNo" example:"P2P20261121123456"`
	ListingID         int64   `json:"listingId" example:"1"`
	ListingType       string  `json:"listingType" enums:"buy,sell" example:"sell"` // buy / sell
	SellerID          int64   `json:"sellerId" example:"1"`
	BuyerID           int64   `json:"buyerId" example:"2"`
	CryptoCurrency    string  `json:"cryptoCurrency" example:"USDT"`
	FiatCurrency      string  `json:"fiatCurrency" example:"TWD"`
	CryptoAmount      float64 `json:"cryptoAmount" example:"100"`
	Price             float64 `json:"price" example:"32.5"`
	FiatAmount        float64 `json:"fiatAmount" example:"3250"`
	PlatformFeeBase   float64 `json:"platformFeeBase" example:"0"`
	PlatformFeeAmount float64 `json:"platformFeeAmount" example:"0"`
	PaymentFeeBase    float64 `json:"paymentFeeBase" example:"0"`
	PaymentFeeAmount  float64 `json:"paymentFeeAmount" example:"0"`
	TotalFee          float64 `json:"totalFee" example:"0"`
	TotalAmount       float64 `json:"totalAmount" example:"3250"`
	PaymentMethodID   int64   `json:"paymentMethodId" example:"1"`
	Status            string  `json:"status" enums:"matched,paid,releasing,completed,cancelled,timeout,disputed,refunded" example:"matched"` // matched / paid / releasing / completed / cancelled / timeout / disputed / refunded
	PaymentDeadline   string  `json:"paymentDeadline" example:"2026-11-21T04:45:00Z"`                                                        // UTC RFC3339
	PaidAt            *string `json:"paidAt,omitempty" example:"2026-11-21T04:35:00Z"`
	ConfirmedAt       *string `json:"confirmedAt,omitempty" example:"2026-11-21T04:40:00Z"`
	CompletedAt       *string `json:"completedAt,omitempty" example:"2026-11-21T04:41:00Z"`
	CancelledAt       *string `json:"cancelledAt,omitempty" example:"2026-11-21T04:36:00Z"`
	CancelReason      *string `json:"cancelReason,omitempty" example:"買方無法付款"`
	CreatedAt         string  `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type ListOrdersResponse struct {
	List []OrderItem `json:"list"`
}
