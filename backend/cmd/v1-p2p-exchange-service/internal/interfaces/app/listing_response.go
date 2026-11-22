package app_interface

import "github.com/shopspring/decimal"

type CreateListingResponse struct {
	ID int64 `json:"id" example:"1"`
}

type ListingItem struct {
	ID               int64           `json:"id" example:"1"`
	UserID           int64           `json:"userId" example:"1"`
	Type             string          `json:"type" example:"sell" enums:"buy,sell"` // buy / sell
	CryptoCurrency   string          `json:"cryptoCurrency" example:"USDT"`
	FiatCurrency     string          `json:"fiatCurrency" example:"TWD"`
	TotalAmount      decimal.Decimal `json:"totalAmount" example:"500"`
	RemainingAmount  decimal.Decimal `json:"remainingAmount" example:"500"`
	Price            decimal.Decimal `json:"price" example:"32.5"`
	MinOrderFiat     decimal.Decimal `json:"minOrderFiat" example:"100"`
	MaxOrderFiat     decimal.Decimal `json:"maxOrderFiat" example:"10000"`
	PlatformFeeBase  decimal.Decimal `json:"platformFeeBase" example:"0"`
	PlatformFeeRate  decimal.Decimal `json:"platformFeeRate" example:"0"`
	PaymentFeeBase   decimal.Decimal `json:"paymentFeeBase" example:"0"`
	PaymentFeeRate   decimal.Decimal `json:"paymentFeeRate" example:"0"`
	PaymentTimeLimit int64           `json:"paymentTimeLimit" example:"15"`
	PaymentMethodID  *int64          `json:"paymentMethodId,omitempty" example:"1"`
	Status           string          `json:"status" example:"active" enums:"active,paused,completed,cancelled"` // active / paused / completed / cancelled
	CreatedAt        string          `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type ListListingsResponse struct {
	List []ListingItem `json:"list"`
}
