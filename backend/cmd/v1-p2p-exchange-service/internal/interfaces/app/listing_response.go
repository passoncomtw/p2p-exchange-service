package app_interface

type CreateListingResponse struct {
	ID int64 `json:"id"`
}

type ListingItem struct {
	ID               int64   `json:"id"`
	UserID           int64   `json:"userId"`
	Type             string  `json:"type"`
	CryptoCurrency   string  `json:"cryptoCurrency"`
	FiatCurrency     string  `json:"fiatCurrency"`
	TotalAmount      float64 `json:"totalAmount"`
	RemainingAmount  float64 `json:"remainingAmount"`
	Price            float64 `json:"price"`
	MinOrderFiat     float64 `json:"minOrderFiat"`
	MaxOrderFiat     float64 `json:"maxOrderFiat"`
	PlatformFeeBase  float64 `json:"platformFeeBase"`
	PlatformFeeRate  float64 `json:"platformFeeRate"`
	PaymentFeeBase   float64 `json:"paymentFeeBase"`
	PaymentFeeRate   float64 `json:"paymentFeeRate"`
	PaymentTimeLimit int64   `json:"paymentTimeLimit"`
	PaymentMethodID  *int64  `json:"paymentMethodId,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"createdAt"`
}

type ListListingsResponse struct {
	List []ListingItem `json:"list"`
}
