package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID                int64           `db:"id"`
	OrderNo           string          `db:"order_no"`
	ListingID         int64           `db:"listing_id"`
	ListingType       string          `db:"listing_type"`
	SellerID          int64           `db:"seller_id"`
	BuyerID           int64           `db:"buyer_id"`
	CryptoCurrency    string          `db:"crypto_currency"`
	FiatCurrency      string          `db:"fiat_currency"`
	CryptoAmount      decimal.Decimal `db:"crypto_amount"`
	Price             decimal.Decimal `db:"price"`
	FiatAmount        decimal.Decimal `db:"fiat_amount"`
	PlatformFeeBase   decimal.Decimal `db:"platform_fee_base"`
	PlatformFeeAmount decimal.Decimal `db:"platform_fee_amount"`
	PaymentFeeBase    decimal.Decimal `db:"payment_fee_base"`
	PaymentFeeAmount  decimal.Decimal `db:"payment_fee_amount"`
	TotalFee          decimal.Decimal `db:"total_fee"`
	TotalAmount       decimal.Decimal `db:"total_amount"`
	PaymentMethodID   int64           `db:"payment_method_id"`
	Status            string          `db:"status"`
	PaymentDeadline   time.Time       `db:"payment_deadline"`
	PaidAt            *time.Time      `db:"paid_at"`
	ConfirmedAt       *time.Time      `db:"confirmed_at"`
	CompletedAt       *time.Time      `db:"completed_at"`
	CancelledAt       *time.Time      `db:"cancelled_at"`
	CancelReason      *string         `db:"cancel_reason"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at"`
}
