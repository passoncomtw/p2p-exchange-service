package app_interface

// ListWalletLedgersRequest 錢包帳本流水查詢條件。
// currency 取自路徑參數，limit/offset 取自 query string。
type ListWalletLedgersRequest struct {
	Currency string `path:"currency" swaggerignore:"true"`
	Limit    int64  `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset   int64  `form:"offset,optional,default=0" default:"0" example:"0"`
}

// WalletItem 單一幣別錢包餘額。
// 餘額欄位為字串（NUMERIC 以 ::text 承接），避免 float 精度誤差。
type WalletItem struct {
	Currency         string `json:"currency" example:"USDT"`
	AvailableBalance string `json:"availableBalance" example:"1000.000000000000000000"` // 字串（NUMERIC(38,18)），避免浮點誤差
	FrozenBalance    string `json:"frozenBalance" example:"0.000000000000000000"`       // 字串（NUMERIC(38,18)），避免浮點誤差
}

type ListWalletsResponse struct {
	List []WalletItem `json:"list"`
}

// WalletLedgerItem 單筆錢包帳本流水。
// Amount 正數為增加、負數為減少；CreatedAt 為 UTC RFC3339。
type WalletLedgerItem struct {
	Type         string  `json:"type" example:"deposit" enums:"deposit,withdraw,freeze,unfreeze,transfer_in,transfer_out,fee_deduct,crypto_deposit,crypto_withdraw,fiat_deposit,fiat_withdraw"` // deposit / withdraw / freeze / unfreeze / transfer_in / transfer_out / fee_deduct / crypto_deposit / crypto_withdraw / fiat_deposit / fiat_withdraw
	Amount       string  `json:"amount" example:"1000.000000000000000000"`                                                                                                                      // 字串；正數為增加、負數為減少
	BalanceAfter string  `json:"balanceAfter" example:"1000.000000000000000000"`
	RefOrderNo   *string `json:"refOrderNo" example:"P2P20261121123456"`   // 關聯訂單編號，無關聯時為 null
	CreatedAt    string  `json:"createdAt" example:"2026-11-21T04:30:00Z"` // UTC RFC3339
}

type ListWalletLedgersResponse struct {
	List  []WalletLedgerItem `json:"list"`
	Total int64              `json:"total" example:"1"`
}
