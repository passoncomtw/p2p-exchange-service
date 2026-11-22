package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Listing struct {
	ID               int64           `db:"id"`
	UserID           int64           `db:"user_id"`
	Type             string          `db:"type"`
	CryptoCurrency   string          `db:"crypto_currency"`
	FiatCurrency     string          `db:"fiat_currency"`
	TotalAmount      decimal.Decimal `db:"total_amount"`
	RemainingAmount  decimal.Decimal `db:"remaining_amount"`
	Price            decimal.Decimal `db:"price"`
	MinOrderFiat     decimal.Decimal `db:"min_order_fiat"`
	MaxOrderFiat     decimal.Decimal `db:"max_order_fiat"`
	PlatformFeeBase  decimal.Decimal `db:"platform_fee_base"`
	PlatformFeeRate  decimal.Decimal `db:"platform_fee_rate"`
	PaymentFeeBase   decimal.Decimal `db:"payment_fee_base"`
	PaymentFeeRate   decimal.Decimal `db:"payment_fee_rate"`
	PaymentTimeLimit int64           `db:"payment_time_limit"`
	PaymentMethodID  *int64          `db:"payment_method_id"`
	Status           string          `db:"status"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
}
