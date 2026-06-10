package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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

type ListingModel struct {
	conn sqlx.SqlConn
}

func NewListingModel(conn sqlx.SqlConn) *ListingModel {
	return &ListingModel{conn: conn}
}

func (m *ListingModel) Create(ctx context.Context, l *Listing) (int64, error) {
	var id int64
	err := m.conn.QueryRowCtx(ctx, &id,
		`INSERT INTO listings (user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, platform_fee_base, platform_fee_rate, payment_fee_base, payment_fee_rate,
		  payment_time_limit, payment_method_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		 RETURNING id`,
		l.UserID, l.Type, l.CryptoCurrency, l.FiatCurrency, l.TotalAmount, l.RemainingAmount, l.Price,
		l.MinOrderFiat, l.MaxOrderFiat, l.PlatformFeeBase, l.PlatformFeeRate, l.PaymentFeeBase, l.PaymentFeeRate,
		l.PaymentTimeLimit, l.PaymentMethodID, l.Status,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *ListingModel) FindByID(ctx context.Context, id int64) (*Listing, error) {
	var l Listing
	err := m.conn.QueryRowCtx(ctx, &l,
		`SELECT id, user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, platform_fee_base, platform_fee_rate, payment_fee_base, payment_fee_rate,
		  payment_time_limit, payment_method_id, status, created_at, updated_at
		 FROM listings WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (m *ListingModel) List(ctx context.Context, listType string, status string, limit, offset int64) ([]*Listing, error) {
	query := `SELECT id, user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, platform_fee_base, platform_fee_rate, payment_fee_base, payment_fee_rate,
		  payment_time_limit, payment_method_id, status, created_at, updated_at
		 FROM listings WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if listType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, listType)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*Listing
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *ListingModel) ListByUser(ctx context.Context, userID int64, listType string, status string, limit, offset int64) ([]*Listing, error) {
	query := `SELECT id, user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, platform_fee_base, platform_fee_rate, payment_fee_base, payment_fee_rate,
		  payment_time_limit, payment_method_id, status, created_at, updated_at
		 FROM listings WHERE user_id = $1`
	args := []interface{}{userID}
	argIdx := 2

	if listType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, listType)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*Listing
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *ListingModel) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE listings SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}

func (m *ListingModel) DeductAmount(ctx context.Context, id int64, amount float64) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE listings SET remaining_amount = remaining_amount - $1, updated_at = NOW()
		 WHERE id = $2 AND remaining_amount >= $1`,
		amount, id,
	)
	return err
}
