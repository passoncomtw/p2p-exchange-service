package app_interface

type CreateOrderRequest struct {
	ListingID   int64   `json:"listingId"`
	CryptoAmount float64 `json:"cryptoAmount"`
}

type GetOrderRequest struct {
	ID int64 `path:"id"`
}

type ListOrdersRequest struct {
	Role   string `form:"role,optional"`
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type UpdateOrderPathRequest struct {
	ID int64 `path:"id"`
}

type CancelOrderRequest struct {
	ID     int64  `path:"id"`
	Reason string `json:"reason"`
}

type DisputeOrderRequest struct {
	ID     int64  `path:"id"`
	Reason string `json:"reason"`
}
