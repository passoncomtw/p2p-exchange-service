package app_interface

// ListWalletLedgersRequest 錢包帳本流水查詢條件。
// currency 取自路徑參數，limit/offset 取自 query string。
type ListWalletLedgersRequest struct {
	Currency string `path:"currency"`
	Limit    int64  `form:"limit,optional,default=20"`
	Offset   int64  `form:"offset,optional,default=0"`
}

// WalletItem 單一幣別錢包餘額。
// 餘額欄位為字串（NUMERIC 以 ::text 承接），避免 float 精度誤差。
type WalletItem struct {
	Currency         string `json:"currency"`
	AvailableBalance string `json:"availableBalance"`
	FrozenBalance    string `json:"frozenBalance"`
}

type ListWalletsResponse struct {
	List []WalletItem `json:"list"`
}

// WalletLedgerItem 單筆錢包帳本流水。
// Amount 正數為增加、負數為減少；CreatedAt 為 UTC RFC3339。
type WalletLedgerItem struct {
	Type         string  `json:"type"`
	Amount       string  `json:"amount"`
	BalanceAfter string  `json:"balanceAfter"`
	RefOrderNo   *string `json:"refOrderNo"`
	CreatedAt    string  `json:"createdAt"`
}

type ListWalletLedgersResponse struct {
	List  []WalletLedgerItem `json:"list"`
	Total int64              `json:"total"`
}
