package app_interface

// CryptoWithdrawRequest 鏈上 USDT 提領申請；提領歸屬人取自 JWT，不接受呼叫端指定。
type CryptoWithdrawRequest struct {
	ToAddress string `json:"toAddress"`
	Amount    string `json:"amount"`
}

type CryptoWithdrawResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type AppListCryptoWithdrawalsRequest struct {
	Limit  int64 `form:"limit,optional,default=20"`
	Offset int64 `form:"offset,optional,default=0"`
}

// CryptoWithdrawalItem 單筆鏈上提領記錄。
// Status: pending（等待廣播）/ broadcasting（已廣播，等待確認）/ confirmed（已確認扣款）/ failed（廣播失敗，餘額已解凍）。
type CryptoWithdrawalItem struct {
	ID          int64   `json:"id"`
	Currency    string  `json:"currency"`
	Amount      string  `json:"amount"`
	ToAddress   string  `json:"toAddress"`
	TxHash      *string `json:"txHash"`
	Status      string  `json:"status"`
	ConfirmedAt *string `json:"confirmedAt"`
	CreatedAt   string  `json:"createdAt"`
}

type AppListCryptoWithdrawalsResponse struct {
	List  []CryptoWithdrawalItem `json:"list"`
	Total int64                  `json:"total"`
}
