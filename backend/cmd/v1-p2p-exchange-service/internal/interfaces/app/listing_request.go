package app_interface

type CreateListingRequest struct {
	Type             string  `json:"type"`
	CryptoCurrency   string  `json:"cryptoCurrency"`
	FiatCurrency     string  `json:"fiatCurrency"`
	TotalAmount      float64 `json:"totalAmount"`
	Price            float64 `json:"price"`
	MinOrderFiat     float64 `json:"minOrderFiat"`
	MaxOrderFiat     float64 `json:"maxOrderFiat"`
	PaymentTimeLimit int64  `json:"paymentTimeLimit"`
	PaymentMethodID  *int64 `json:"paymentMethodId,omitempty"`
}

type ListListingsRequest struct {
	Type   string `form:"type,optional"`
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type GetListingRequest struct {
	ID int64 `path:"id"`
}

type CancelListingRequest struct {
	ID int64 `path:"id"`
}
