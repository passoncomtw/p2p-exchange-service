package app_interface

// FiatWithdrawRequest 建立 TWD 提領申請的請求。
// 金額以字串傳遞（避免浮點誤差），由 service 驗證格式與上下限。
type FiatWithdrawRequest struct {
	Amount      string `json:"amount" example:"1000"`              // 字串，範圍 100 ~ 100000 TWD
	BankCode    string `json:"bankCode" example:"004"`             // 3 位數字
	BankAccount string `json:"bankAccount" example:"123456789012"` // 6~20 位數字
	AccountName string `json:"accountName" example:"王小明"`
}

// FiatWithdrawResponse 提領申請建立結果（狀態固定為 pending，等待後台審核）。
type FiatWithdrawResponse struct {
	ID     int64  `json:"id" example:"1"`
	Status string `json:"status" example:"pending"` // 固定為 pending，等待後台審核
}

type AppListFiatWithdrawalsRequest struct {
	Limit  int64 `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64 `form:"offset,optional,default=0" default:"0" example:"0"`
}

// AppFiatWithdrawalItem 單筆提領記錄。
// Status: pending（等待審核）/ approved（已核可扣款）/ rejected（已拒絕並解凍）。
// BankAccountTail 為遮蔽後的銀行帳號（僅保留後 4 碼），不回傳完整帳號。
type AppFiatWithdrawalItem struct {
	ID              int64  `json:"id" example:"1"`
	Currency        string `json:"currency" example:"TWD"`
	Amount          string `json:"amount" example:"1000.00"`
	BankCode        string `json:"bankCode" example:"004"`
	BankAccountTail string `json:"bankAccountTail" example:"9012"`                             // 遮蔽後的帳號，僅保留後 4 碼
	Status          string `json:"status" example:"pending" enums:"pending,approved,rejected"` // pending / approved / rejected
	CreatedAt       string `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type AppListFiatWithdrawalsResponse struct {
	List  []AppFiatWithdrawalItem `json:"list"`
	Total int64                   `json:"total" example:"1"`
}
