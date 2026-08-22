package app_interface

// FiatWithdrawRequest 建立 TWD 提領申請的請求。
// 金額以字串傳遞（避免浮點誤差），由 service 驗證格式與上下限。
type FiatWithdrawRequest struct {
	Amount      string `json:"amount"`
	BankCode    string `json:"bankCode"`
	BankAccount string `json:"bankAccount"`
	AccountName string `json:"accountName"`
}

// FiatWithdrawResponse 提領申請建立結果（狀態固定為 pending，等待後台審核）。
type FiatWithdrawResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type AppListFiatWithdrawalsRequest struct {
	Limit  int64 `form:"limit,optional,default=20"`
	Offset int64 `form:"offset,optional,default=0"`
}

// AppFiatWithdrawalItem 單筆提領記錄。
// Status: pending（等待審核）/ approved（已核可扣款）/ rejected（已拒絕並解凍）。
// BankAccountTail 為遮蔽後的銀行帳號（僅保留後 4 碼），不回傳完整帳號。
type AppFiatWithdrawalItem struct {
	ID              int64  `json:"id"`
	Currency        string `json:"currency"`
	Amount          string `json:"amount"`
	BankCode        string `json:"bankCode"`
	BankAccountTail string `json:"bankAccountTail"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
}

type AppListFiatWithdrawalsResponse struct {
	List  []AppFiatWithdrawalItem `json:"list"`
	Total int64                   `json:"total"`
}
