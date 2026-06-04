package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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

type OrderModel struct {
	conn sqlx.SqlConn
}

func NewOrderModel(conn sqlx.SqlConn) *OrderModel {
	return &OrderModel{conn: conn}
}

func (m *OrderModel) Create(ctx context.Context, o *Order) (int64, error) {
	var id int64
	err := m.conn.QueryRowCtx(ctx, &id,
		`INSERT INTO orders (order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
		  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
		  total_fee, total_amount, payment_method_id, status, payment_deadline, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW(), NOW())
		 RETURNING id`,
		o.OrderNo, o.ListingID, o.ListingType, o.SellerID, o.BuyerID, o.CryptoCurrency, o.FiatCurrency,
		o.CryptoAmount, o.Price, o.FiatAmount, o.PlatformFeeBase, o.PlatformFeeAmount, o.PaymentFeeBase, o.PaymentFeeAmount,
		o.TotalFee, o.TotalAmount, o.PaymentMethodID, o.Status, o.PaymentDeadline,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *OrderModel) FindByID(ctx context.Context, id int64) (*Order, error) {
	var o Order
	err := m.conn.QueryRowCtx(ctx, &o,
		`SELECT id, order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
		  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
		  total_fee, total_amount, payment_method_id, status, payment_deadline, paid_at, confirmed_at, completed_at,
		  cancelled_at, cancel_reason, created_at, updated_at
		 FROM orders WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (m *OrderModel) FindByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	var o Order
	err := m.conn.QueryRowCtx(ctx, &o,
		`SELECT id, order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
		  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
		  total_fee, total_amount, payment_method_id, status, payment_deadline, paid_at, confirmed_at, completed_at,
		  cancelled_at, cancel_reason, created_at, updated_at
		 FROM orders WHERE order_no = $1`,
		orderNo,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (m *OrderModel) List(ctx context.Context, userID int64, role string, status string, limit, offset int64) ([]*Order, error) {
	query := `SELECT id, order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
		  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
		  total_fee, total_amount, payment_method_id, status, payment_deadline, paid_at, confirmed_at, completed_at,
		  cancelled_at, cancel_reason, created_at, updated_at
		 FROM orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if userID != 0 {
		switch role {
		case "buyer":
			query += fmt.Sprintf(" AND buyer_id = $%d", argIdx)
			args = append(args, userID)
			argIdx++
		case "seller":
			query += fmt.Sprintf(" AND seller_id = $%d", argIdx)
			args = append(args, userID)
			argIdx++
		}
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*Order
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *OrderModel) UpdateStatus(ctx context.Context, id int64, status string, extras map[string]interface{}) error {
	query := "UPDATE orders SET status = $1, updated_at = NOW()"
	args := []interface{}{status}
	argIdx := 2

	for key, val := range extras {
		switch key {
		case "paid_at":
			query += fmt.Sprintf(", paid_at = $%d", argIdx)
			args = append(args, val)
			argIdx++
		case "confirmed_at":
			query += fmt.Sprintf(", confirmed_at = $%d", argIdx)
			args = append(args, val)
			argIdx++
		case "completed_at":
			query += fmt.Sprintf(", completed_at = $%d", argIdx)
			args = append(args, val)
			argIdx++
		case "cancelled_at":
			query += fmt.Sprintf(", cancelled_at = $%d", argIdx)
			args = append(args, val)
			argIdx++
		case "cancel_reason":
			query += fmt.Sprintf(", cancel_reason = $%d", argIdx)
			args = append(args, val)
			argIdx++
		}
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}
