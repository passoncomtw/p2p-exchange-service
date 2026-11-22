package app_interface

// CryptoWithdrawRequest 鏈上 USDT 提領申請；提領歸屬人取自 JWT，不接受呼叫端指定。
type CryptoWithdrawRequest struct {
	ToAddress string `json:"toAddress" example:"TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"` // Tron Base58Check 地址
	Amount    string `json:"amount" example:"10"`                                    // 字串，十進位正數（最多 18 位小數），最低 10 USDT
}

type CryptoWithdrawResponse struct {
	ID     int64  `json:"id" example:"1"`
	Status string `json:"status" example:"pending"` // pending
}

type AppListCryptoWithdrawalsRequest struct {
	Limit  int64 `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64 `form:"offset,optional,default=0" default:"0" example:"0"`
}

// CryptoWithdrawalItem 單筆鏈上提領記錄。
// Status: pending（等待廣播）/ broadcasting（已廣播，等待確認）/ confirmed（已確認扣款）/ failed（廣播失敗，餘額已解凍）。
type CryptoWithdrawalItem struct {
	ID          int64   `json:"id" example:"1"`
	Currency    string  `json:"currency" example:"USDT"`
	Amount      string  `json:"amount" example:"10.000000000000000000"`
	ToAddress   string  `json:"toAddress" example:"TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"`
	TxHash      *string `json:"txHash" example:"9f2c1b7e5a3d4c8b6e0f1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f"` // 尚未廣播時為 null
	Status      string  `json:"status" example:"confirmed" enums:"pending,broadcasting,confirmed,failed"`          // pending / broadcasting / confirmed / failed
	ConfirmedAt *string `json:"confirmedAt" example:"2026-11-21T04:30:00Z"`                                        // 尚未確認時為 null
	CreatedAt   string  `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type AppListCryptoWithdrawalsResponse struct {
	List  []CryptoWithdrawalItem `json:"list"`
	Total int64                  `json:"total" example:"1"`
}
