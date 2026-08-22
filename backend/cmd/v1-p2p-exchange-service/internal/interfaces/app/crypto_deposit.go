package app_interface

// GetCryptoDepositInfoResponse USDT 充值資訊：使用者須將 USDT 轉入 Address，
// 並在鏈上交易的 memo 欄位帶入 Memo（8 位小寫十六進位的 userID）才能自動入帳。
type GetCryptoDepositInfoResponse struct {
	Address         string `json:"address"`
	Memo            string `json:"memo"`
	Network         string `json:"network"`
	Currency        string `json:"currency"`
	ContractAddress string `json:"contractAddress"`
}

type AppListCryptoDepositsRequest struct {
	Limit  int64 `form:"limit,optional,default=20"`
	Offset int64 `form:"offset,optional,default=0"`
}

// CryptoDepositItem 單筆鏈上充值記錄。
// Status: pending（等待區塊確認）/ confirmed（已入帳）/ failed（memo 無法識別或確認超時）。
type CryptoDepositItem struct {
	ID          int64   `json:"id"`
	Currency    string  `json:"currency"`
	Amount      string  `json:"amount"`
	TxHash      string  `json:"txHash"`
	Status      string  `json:"status"`
	ConfirmedAt *string `json:"confirmedAt"`
	CreatedAt   string  `json:"createdAt"`
}

type AppListCryptoDepositsResponse struct {
	List  []CryptoDepositItem `json:"list"`
	Total int64               `json:"total"`
}
