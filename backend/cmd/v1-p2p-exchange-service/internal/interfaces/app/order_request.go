package app_interface

import "p2p-exchange/internal/money"

type CreateOrderRequest struct {
	ListingID    int64        `json:"listingId" example:"1"`
	CryptoAmount money.Amount `json:"cryptoAmount" example:"100"` // 必須為正數，最多 18 位小數；換算法幣須落在掛單 minOrderFiat ~ maxOrderFiat
}

type GetOrderRequest struct {
	ID int64 `path:"id" swaggerignore:"true"`
}

type ListOrdersRequest struct {
	Role   string `form:"role,optional" enums:"buyer,seller" example:"buyer"`
	Status string `form:"status,optional" enums:"matched,paid,releasing,completed,cancelled,timeout,disputed,refunded" example:"matched"`
	Limit  int64  `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

type UpdateOrderPathRequest struct {
	ID int64 `path:"id" swaggerignore:"true"`
}

type CancelOrderRequest struct {
	ID     int64  `path:"id" swaggerignore:"true"`
	Reason string `json:"reason" example:"買方無法付款"` // 僅 matched 狀態可取消
}

type DisputeOrderRequest struct {
	ID     int64  `path:"id" swaggerignore:"true"`
	Reason string `json:"reason" example:"賣方已付款但買方未放行"` // 僅 paid 狀態可申訴
}
