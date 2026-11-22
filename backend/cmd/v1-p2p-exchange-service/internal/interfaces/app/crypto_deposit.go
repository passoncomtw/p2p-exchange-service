package app_interface

// GetCryptoDepositInfoResponse USDT 充值資訊：使用者須將 USDT 轉入 Address，
// 並在鏈上交易的 memo 欄位帶入 Memo（8 位小寫十六進位的 userID）才能自動入帳。
type GetCryptoDepositInfoResponse struct {
	Address         string `json:"address" example:"TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"`
	Memo            string `json:"memo" example:"00000001"` // 8 位小寫十六進位的 userID；轉帳時必須帶入才能自動入帳
	Network         string `json:"network" example:"nile"`
	Currency        string `json:"currency" example:"USDT"`
	ContractAddress string `json:"contractAddress" example:"TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj"`
}

type AppListCryptoDepositsRequest struct {
	Limit  int64 `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64 `form:"offset,optional,default=0" default:"0" example:"0"`
}

// CryptoDepositItem 單筆鏈上充值記錄。
// Status: pending（等待區塊確認）/ confirmed（已入帳）/ failed（memo 無法識別或確認超時）。
type CryptoDepositItem struct {
	ID          int64   `json:"id" example:"1"`
	Currency    string  `json:"currency" example:"USDT"`
	Amount      string  `json:"amount" example:"100.000000000000000000"`
	TxHash      string  `json:"txHash" example:"9f2c1b7e5a3d4c8b6e0f1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f"`
	Status      string  `json:"status" example:"confirmed" enums:"pending,confirmed,failed"` // pending / confirmed / failed
	ConfirmedAt *string `json:"confirmedAt" example:"2026-11-21T04:30:00Z"`                  // 尚未確認時為 null
	CreatedAt   string  `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type AppListCryptoDepositsResponse struct {
	List  []CryptoDepositItem `json:"list"`
	Total int64               `json:"total" example:"1"`
}
