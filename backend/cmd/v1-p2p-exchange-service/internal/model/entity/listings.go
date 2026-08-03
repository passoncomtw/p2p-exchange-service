package entity

import "time"

type Listing struct {
	ID               int64     `db:"id"`
	UserID           int64     `db:"user_id"`
	Type             string    `db:"type"`
	CryptoCurrency   string    `db:"crypto_currency"`
	FiatCurrency     string    `db:"fiat_currency"`
	TotalAmount      float64   `db:"total_amount"`
	RemainingAmount  float64   `db:"remaining_amount"`
	Price            float64   `db:"price"`
	MinOrderFiat     float64   `db:"min_order_fiat"`
	MaxOrderFiat     float64   `db:"max_order_fiat"`
	PlatformFeeBase  float64   `db:"platform_fee_base"`
	PlatformFeeRate  float64   `db:"platform_fee_rate"`
	PaymentFeeBase   float64   `db:"payment_fee_base"`
	PaymentFeeRate   float64   `db:"payment_fee_rate"`
	PaymentTimeLimit int64     `db:"payment_time_limit"`
	PaymentMethodID  *int64    `db:"payment_method_id"`
	Status           string    `db:"status"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
