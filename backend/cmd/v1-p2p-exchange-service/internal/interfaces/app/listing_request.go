package app_interface

import "p2p-exchange/internal/money"

type CreateListingRequest struct {
	Type             string       `json:"type" enums:"buy,sell" example:"sell"` // buy / sell；sell 必須帶 paymentMethodId，且會凍結 totalAmount
	CryptoCurrency   string       `json:"cryptoCurrency" example:"USDT"`
	FiatCurrency     string       `json:"fiatCurrency" example:"TWD"`
	TotalAmount      money.Amount `json:"totalAmount" example:"500"`     // 必須大於 0
	Price            money.Amount `json:"price" example:"32.5"`          // 必須大於 0
	MinOrderFiat     money.Amount `json:"minOrderFiat" example:"100"`    // 不可為負
	MaxOrderFiat     money.Amount `json:"maxOrderFiat" example:"10000"`  // 不可小於 minOrderFiat
	PaymentTimeLimit int64        `json:"paymentTimeLimit" example:"15"` // 付款時限（分鐘），必須大於 0
	PaymentMethodID  *int64       `json:"paymentMethodId,omitempty" example:"1"`
}

type ListListingsRequest struct {
	Type   string `form:"type,optional" example:"sell" enums:"buy,sell"`
	Status string `form:"status,optional" example:"active" enums:"active,paused,completed,cancelled"`
	Limit  int64  `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

type GetListingRequest struct {
	ID int64 `path:"id" swaggerignore:"true"`
}

type CancelListingRequest struct {
	ID int64 `path:"id" swaggerignore:"true"`
}
