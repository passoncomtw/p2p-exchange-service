package app_interface

// FiatDepositRequest 建立 TWD 入金訂單的請求（金額以整數元計）。
type FiatDepositRequest struct {
	Amount int `json:"amount"`
}

// FiatDepositResponse ECPay AioCheckOut 導轉資訊。
// FormParams 已含 CheckMacValue，前端以 POST form 送往 PaymentURL。
type FiatDepositResponse struct {
	MerchantTradeNo string            `json:"merchantTradeNo"`
	PaymentURL      string            `json:"paymentUrl"`
	FormParams      map[string]string `json:"formParams"`
}

type AppListFiatDepositsRequest struct {
	Limit  int64 `form:"limit,optional,default=20"`
	Offset int64 `form:"offset,optional,default=0"`
}

// FiatDepositItem 單筆入金記錄。
// Status: pending（等待付款）/ paid（已付款）/ failed（付款失敗）。
type FiatDepositItem struct {
	ID              int64   `json:"id"`
	Currency        string  `json:"currency"`
	Amount          string  `json:"amount"`
	MerchantTradeNo string  `json:"merchantTradeNo"`
	Status          string  `json:"status"`
	PaymentType     *string `json:"paymentType"`
	PaidAt          *string `json:"paidAt"`
	CreatedAt       string  `json:"createdAt"`
}

type AppListFiatDepositsResponse struct {
	List  []FiatDepositItem `json:"list"`
	Total int64             `json:"total"`
}
