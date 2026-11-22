package app_interface

// FiatDepositRequest 建立 TWD 入金訂單的請求（金額以整數元計）。
type FiatDepositRequest struct {
	Amount int `json:"amount" example:"1000"` // 整數 TWD，範圍 100 ~ 50000
}

// FiatDepositResponse ECPay AioCheckOut 導轉資訊。
// FormParams 已含 CheckMacValue，前端以 POST form 送往 PaymentURL。
type FiatDepositResponse struct {
	MerchantTradeNo string            `json:"merchantTradeNo" example:"P178495360000000"`
	PaymentURL      string            `json:"paymentUrl" example:"https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5"`
	FormParams      map[string]string `json:"formParams" example:"MerchantID:2000132,TotalAmount:1000"` // 已含 CheckMacValue，前端以 POST form 送往 paymentUrl
}

type AppListFiatDepositsRequest struct {
	Limit  int64 `form:"limit,optional,default=20" default:"20" example:"20"`
	Offset int64 `form:"offset,optional,default=0" default:"0" example:"0"`
}

// FiatDepositItem 單筆入金記錄。
// Status: pending（等待付款）/ paid（已付款）/ failed（付款失敗）。
type FiatDepositItem struct {
	ID              int64   `json:"id" example:"1"`
	Currency        string  `json:"currency" example:"TWD"`
	Amount          string  `json:"amount" example:"1000.00"`
	MerchantTradeNo string  `json:"merchantTradeNo" example:"P178495360000000"`
	Status          string  `json:"status" example:"paid" enums:"pending,paid,failed"` // pending / paid / failed
	PaymentType     *string `json:"paymentType" example:"Credit_CreditCard"`           // 尚未付款時為 null
	PaidAt          *string `json:"paidAt" example:"2026-11-21T04:30:00Z"`             // 尚未付款時為 null
	CreatedAt       string  `json:"createdAt" example:"2026-11-21T04:30:00Z"`
}

type AppListFiatDepositsResponse struct {
	List  []FiatDepositItem `json:"list"`
	Total int64             `json:"total" example:"1"`
}
