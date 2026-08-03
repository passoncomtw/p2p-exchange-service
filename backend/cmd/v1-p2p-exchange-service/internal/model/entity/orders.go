package entity

import "time"

type Order struct {
	ID                int64      `db:"id"`
	OrderNo           string     `db:"order_no"`
	ListingID         int64      `db:"listing_id"`
	ListingType       string     `db:"listing_type"`
	SellerID          int64      `db:"seller_id"`
	BuyerID           int64      `db:"buyer_id"`
	CryptoCurrency    string     `db:"crypto_currency"`
	FiatCurrency      string     `db:"fiat_currency"`
	CryptoAmount      float64    `db:"crypto_amount"`
	Price             float64    `db:"price"`
	FiatAmount        float64    `db:"fiat_amount"`
	PlatformFeeBase   float64    `db:"platform_fee_base"`
	PlatformFeeAmount float64    `db:"platform_fee_amount"`
	PaymentFeeBase    float64    `db:"payment_fee_base"`
	PaymentFeeAmount  float64    `db:"payment_fee_amount"`
	TotalFee          float64    `db:"total_fee"`
	TotalAmount       float64    `db:"total_amount"`
	PaymentMethodID   int64      `db:"payment_method_id"`
	Status            string     `db:"status"`
	PaymentDeadline   time.Time  `db:"payment_deadline"`
	PaidAt            *time.Time `db:"paid_at"`
	ConfirmedAt       *time.Time `db:"confirmed_at"`
	CompletedAt       *time.Time `db:"completed_at"`
	CancelledAt       *time.Time `db:"cancelled_at"`
	CancelReason      *string    `db:"cancel_reason"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}
