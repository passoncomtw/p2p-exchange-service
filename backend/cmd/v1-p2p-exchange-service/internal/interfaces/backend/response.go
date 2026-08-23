package backend_interface

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

type DashboardResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// MemberItem 後台會員列表單筆資料（Email 為 NULL 時輸出空字串）。
type MemberItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type BackendListMembersResponse struct {
	List  []MemberItem `json:"list"`
	Total int64        `json:"total"`
}

// BackendDepositResponse 手動入金後的錢包餘額快照。
type BackendDepositResponse struct {
	Currency         string `json:"currency"`
	AvailableBalance string `json:"available_balance"`
	FrozenBalance    string `json:"frozen_balance"`
}

// TronWalletInfo 平台 Tron 熱錢包資訊。
// OnChainBalance 為鏈上實際餘額，PlatformVirtualBalance 為 DB 內所有使用者 USDT 餘額總和；
// 任一來源查詢失敗時該欄位為 "0"（不中斷回應）。
type TronWalletInfo struct {
	Enabled                bool   `json:"enabled"`
	Network                string `json:"network"`
	HotWalletAddress       string `json:"hotWalletAddress"`
	USDTContractAddress    string `json:"usdtContractAddress"`
	ConfirmationBlocks     int    `json:"confirmationBlocks"`
	OnChainBalance         string `json:"onChainBalance"`
	PlatformVirtualBalance string `json:"platformVirtualBalance"`
}

// ECPayInfo 平台 ECPay 金流設定資訊（僅輸出非機敏欄位，不含 HashKey / HashIV）。
type ECPayInfo struct {
	Enabled    bool   `json:"enabled"`
	MerchantID string `json:"merchantId"`
	BaseURL    string `json:"baseUrl"`
}

type PlatformWalletInfoResponse struct {
	Tron  TronWalletInfo `json:"tron"`
	ECPay ECPayInfo      `json:"ecpay"`
}

// FiatWithdrawalItem 後台法幣提領申請列表單筆資料。
// BankAccount 為完整帳號（後台審核人員需核對匯款資訊，刻意不遮蔽），
// 與 App 端列表（只回後 4 碼）不同；此結構只會經 Backend JWT 保護的路由輸出。
type FiatWithdrawalItem struct {
	ID           int64   `json:"id"`
	UserID       int64   `json:"userId"`
	Currency     string  `json:"currency"`
	Amount       string  `json:"amount"`
	BankCode     string  `json:"bankCode"`
	BankAccount  string  `json:"bankAccount"`
	AccountName  string  `json:"accountName"`
	Status       string  `json:"status"`
	ReviewedBy   *int64  `json:"reviewedBy"`
	RejectReason *string `json:"rejectReason"`
	CreatedAt    string  `json:"createdAt"`
}

type BackendListFiatWithdrawalsResponse struct {
	List  []FiatWithdrawalItem `json:"list"`
	Total int64                `json:"total"`
}
