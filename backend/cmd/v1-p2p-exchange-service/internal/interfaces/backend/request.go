package backend_interface

type BackendLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BackendListListingsRequest struct {
	Type   string `form:"type,optional"`
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type BackendListOrdersRequest struct {
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type BackendResolveOrderRequest struct {
	ID     int64  `path:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}
