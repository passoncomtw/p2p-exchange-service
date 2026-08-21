package entity

import "time"

// FiatDeposit 對應 fiat_deposits 表（ECPay TWD 入金記錄）。
// MerchantTradeNo 為平台產生的唯一交易號（送給 ECPay）；EcpayOrderNo 為 ECPay 回傳的訂單編號。
// Status: pending（等待付款）/ paid（Webhook 確認付款成功）/ failed（付款失敗或 RtnCode != 1）。
type FiatDeposit struct {
	ID              int64      `db:"id"`
	UserID          int64      `db:"user_id"`
	Currency        string     `db:"currency"`
	Amount          string     `db:"amount"`
	EcpayOrderNo    *string    `db:"ecpay_order_no"`
	MerchantTradeNo string     `db:"merchant_trade_no"`
	Status          string     `db:"status"`
	PaymentType     *string    `db:"payment_type"`
	PaidAt          *time.Time `db:"paid_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
