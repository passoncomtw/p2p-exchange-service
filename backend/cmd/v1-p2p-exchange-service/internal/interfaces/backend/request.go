package backend_interface

type LoginRequest struct {
	Username string `json:"username" example:"testdemo001"` // App 登入帳號；後台登入請用 admin001
	Password string `json:"password" example:"a12345678"`
}

type BackendListListingsRequest struct {
	Type   string `form:"type,optional" example:"sell" enums:"buy,sell"`
	Status string `form:"status,optional" example:"active" enums:"active,paused,completed,cancelled"`
	Limit  int64  `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

type BackendListOrdersRequest struct {
	Status string `form:"status,optional" example:"disputed" enums:"matched,paid,releasing,completed,cancelled,timeout,disputed,refunded"`
	Limit  int64  `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

type ResolveOrderRequest struct {
	ID     int64  `path:"id" swaggerignore:"true"`
	Action string `json:"action" example:"complete" enums:"complete,refund"` // complete（放行給買方）/ refund（退回賣方，狀態轉 refunded）；僅 disputed 狀態可處理
	Reason string `json:"reason" example:"經查證買方已付款，放行給買方"`
}

type BackendListMembersRequest struct {
	Keyword string `form:"keyword,optional" example:"testdemo"`
	Limit   int64  `form:"limit,optional,default=10" default:"20" example:"20"`
	Offset  int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

// BackendDepositRequest 後台手動入金請求。
// ID 為入金對象（路徑參數）；操作人一律取自 JWT context，刻意不在此結構出現，
// 避免呼叫端從 body 偽造操作人。
type BackendDepositRequest struct {
	ID       int64  `path:"id" swaggerignore:"true"`
	Currency string `json:"currency" example:"USDT"` // 必須為 currencies 表內存在的幣別
	Amount   string `json:"amount" example:"100"`    // 字串，必須為正數
}

// BackendListFiatWithdrawalsRequest 後台法幣提領申請列表查詢條件。
// Status 為空字串或 "all" 時回全部狀態。
type BackendListFiatWithdrawalsRequest struct {
	Status string `form:"status,optional,default=pending" example:"pending" enums:"pending,approved,rejected,all"`
	Limit  int64  `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

// BackendReviewFiatWithdrawalRequest 後台法幣提領審核請求。
// ID 為被審核的申請（路徑參數）；審核人一律取自 Backend JWT context，刻意不在此結構出現，
// 避免呼叫端從 body 偽造審核人。
type BackendReviewFiatWithdrawalRequest struct {
	ID     int64  `path:"id" swaggerignore:"true"`
	Action string `json:"action" example:"approve" enums:"approve,reject"` // approve | reject
	Reason string `json:"reason,optional" example:"帳戶名稱與會員不符"`             // action 為 reject 時必填
}
