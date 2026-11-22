package backend_interface

type LoginResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTc5NTI0MjYwMH0.abc123"`
	ExpiresIn int64  `json:"expiresIn" example:"86400"`
}

type DashboardResponse struct {
	Username string `json:"username" example:"admin001"`
	Role     string `json:"role" example:"admin"`
}

// MemberItem 後台會員列表單筆資料（Email 為 NULL 時輸出空字串）。
type MemberItem struct {
	ID        int64  `json:"id" example:"1"`
	Username  string `json:"username" example:"testdemo001"`
	Email     string `json:"email" example:"testdemo001@example.com"`
	CreatedAt string `json:"createdAt" example:"2026-11-21T04:30:00Z"`
	UpdatedAt string `json:"updatedAt" example:"2026-11-21T04:30:00Z"`
}

type BackendListMembersResponse struct {
	List  []MemberItem `json:"list"`
	Total int64        `json:"total" example:"1"`
}

// BackendDepositResponse 手動入金後的錢包餘額快照。
type BackendDepositResponse struct {
	Currency         string `json:"currency" example:"USDT"`
	AvailableBalance string `json:"available_balance" example:"1100.000000000000000000"`
	FrozenBalance    string `json:"frozen_balance" example:"0.000000000000000000"`
}

// TronWalletInfo 平台 Tron 熱錢包資訊。
// OnChainBalance 為鏈上實際餘額，PlatformVirtualBalance 為 DB 內所有使用者 USDT 餘額總和；
// 任一來源查詢失敗時該欄位為 "0"（不中斷回應）。
type TronWalletInfo struct {
	Enabled                bool   `json:"enabled" example:"true"`
	Network                string `json:"network" example:"nile"`
	HotWalletAddress       string `json:"hotWalletAddress" example:"TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"`
	USDTContractAddress    string `json:"usdtContractAddress" example:"TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj"`
	ConfirmationBlocks     int    `json:"confirmationBlocks" example:"6"`
	OnChainBalance         string `json:"onChainBalance" example:"5000.000000"`                     // 查詢失敗時為 "0"
	PlatformVirtualBalance string `json:"platformVirtualBalance" example:"4800.000000000000000000"` // 查詢失敗時為 "0"
}

// ECPayInfo 平台 ECPay 金流設定資訊（僅輸出非機敏欄位，不含 HashKey / HashIV）。
type ECPayInfo struct {
	Enabled    bool   `json:"enabled" example:"true"`
	MerchantID string `json:"merchantId" example:"2000132"`
	BaseURL    string `json:"baseUrl" example:"https://payment-stage.ecpay.com.tw"`
}

type PlatformWalletInfoResponse struct {
	Tron  TronWalletInfo `json:"tron"`
	ECPay ECPayInfo      `json:"ecpay"`
}

// FiatWithdrawalItem 後台法幣提領申請列表單筆資料。
// BankAccount 為完整帳號（後台審核人員需核對匯款資訊，刻意不遮蔽），
// 與 App 端列表（只回後 4 碼）不同；此結構只會經 Backend JWT 保護的路由輸出。
type FiatWithdrawalItem struct {
	ID           int64   `json:"id" example:"1"`
	UserID       int64   `json:"userId" example:"1"`
	Currency     string  `json:"currency" example:"TWD"`
	Amount       string  `json:"amount" example:"1000.00"`
	BankCode     string  `json:"bankCode" example:"004"`
	BankAccount  string  `json:"bankAccount" example:"123456789012"`
	AccountName  string  `json:"accountName" example:"王小明"`
	Status       string  `json:"status" example:"pending" enums:"pending,approved,rejected"` // pending / approved / rejected
	ReviewedBy   *int64  `json:"reviewedBy" example:"1"`                                     // 尚未審核時為 null
	RejectReason *string `json:"rejectReason" example:"帳戶名稱與會員不符"`                           // 未駁回時為 null
	CreatedAt    string  `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type BackendListFiatWithdrawalsResponse struct {
	List  []FiatWithdrawalItem `json:"list"`
	Total int64                `json:"total" example:"1"`
}
