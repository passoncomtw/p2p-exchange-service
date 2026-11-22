package app_interface

type CreateListingResponse struct {
	ID int64 `json:"id" example:"1"`
}

type ListingItem struct {
	ID               int64   `json:"id" example:"1"`
	UserID           int64   `json:"userId" example:"1"`
	Type             string  `json:"type" example:"sell" enums:"buy,sell"` // buy / sell
	CryptoCurrency   string  `json:"cryptoCurrency" example:"USDT"`
	FiatCurrency     string  `json:"fiatCurrency" example:"TWD"`
	TotalAmount      float64 `json:"totalAmount" example:"500"`
	RemainingAmount  float64 `json:"remainingAmount" example:"500"`
	Price            float64 `json:"price" example:"32.5"`
	MinOrderFiat     float64 `json:"minOrderFiat" example:"100"`
	MaxOrderFiat     float64 `json:"maxOrderFiat" example:"10000"`
	PlatformFeeBase  float64 `json:"platformFeeBase" example:"0"`
	PlatformFeeRate  float64 `json:"platformFeeRate" example:"0"`
	PaymentFeeBase   float64 `json:"paymentFeeBase" example:"0"`
	PaymentFeeRate   float64 `json:"paymentFeeRate" example:"0"`
	PaymentTimeLimit int64   `json:"paymentTimeLimit" example:"15"`
	PaymentMethodID  *int64  `json:"paymentMethodId,omitempty" example:"1"`
	Status           string  `json:"status" example:"active" enums:"active,paused,completed,cancelled"` // active / paused / completed / cancelled
	CreatedAt        string  `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type ListListingsResponse struct {
	List []ListingItem `json:"list"`
}
