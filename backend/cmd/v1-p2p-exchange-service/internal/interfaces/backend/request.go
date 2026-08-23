package backend_interface

type LoginRequest struct {
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

type ResolveOrderRequest struct {
	ID     int64  `path:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type BackendListMembersRequest struct {
	Keyword string `form:"keyword,optional"`
	Limit   int64  `form:"limit,optional,default=10"`
	Offset  int64  `form:"offset,optional,default=0"`
}

// BackendDepositRequest 後台手動入金請求。
// ID 為入金對象（路徑參數）；操作人一律取自 JWT context，刻意不在此結構出現，
// 避免呼叫端從 body 偽造操作人。
type BackendDepositRequest struct {
	ID       int64  `path:"id"`
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

// BackendListFiatWithdrawalsRequest 後台法幣提領申請列表查詢條件。
// Status 為空字串或 "all" 時回全部狀態。
type BackendListFiatWithdrawalsRequest struct {
	Status string `form:"status,optional,default=pending"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

// BackendReviewFiatWithdrawalRequest 後台法幣提領審核請求。
// ID 為被審核的申請（路徑參數）；審核人一律取自 Backend JWT context，刻意不在此結構出現，
// 避免呼叫端從 body 偽造審核人。
type BackendReviewFiatWithdrawalRequest struct {
	ID     int64  `path:"id"`
	Action string `json:"action"` // approve | reject
	Reason string `json:"reason,optional"`
}
