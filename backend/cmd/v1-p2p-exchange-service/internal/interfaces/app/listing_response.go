package app_interface

import "github.com/shopspring/decimal"

type CreateListingResponse struct {
	ID int64 `json:"id"`
}

type ListingItem struct {
	ID               int64           `json:"id"`
	UserID           int64           `json:"userId"`
	Type             string          `json:"type"`
	CryptoCurrency   string          `json:"cryptoCurrency"`
	FiatCurrency     string          `json:"fiatCurrency"`
	TotalAmount      decimal.Decimal `json:"totalAmount"`
	RemainingAmount  decimal.Decimal `json:"remainingAmount"`
	Price            decimal.Decimal `json:"price"`
	MinOrderFiat     decimal.Decimal `json:"minOrderFiat"`
	MaxOrderFiat     decimal.Decimal `json:"maxOrderFiat"`
	PlatformFeeBase  decimal.Decimal `json:"platformFeeBase"`
	PlatformFeeRate  decimal.Decimal `json:"platformFeeRate"`
	PaymentFeeBase   decimal.Decimal `json:"paymentFeeBase"`
	PaymentFeeRate   decimal.Decimal `json:"paymentFeeRate"`
	PaymentTimeLimit int64           `json:"paymentTimeLimit"`
	PaymentMethodID  *int64          `json:"paymentMethodId,omitempty"`
	Status           string          `json:"status"`
	CreatedAt        string          `json:"createdAt"`
}

type ListListingsResponse struct {
	List []ListingItem `json:"list"`
}
