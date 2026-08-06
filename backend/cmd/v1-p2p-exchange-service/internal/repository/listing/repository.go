package listingrepo

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

const listingSelectCols = `id, user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
  min_order_fiat, max_order_fiat, platform_fee_base, platform_fee_rate, payment_fee_base, payment_fee_rate,
  payment_time_limit, payment_method_id, status, created_at, updated_at`

type ListingRepository interface {
	Create(ctx context.Context, l *entity.Listing) (int64, error)
	FindByID(ctx context.Context, id int64) (*entity.Listing, error)
	List(ctx context.Context, listType, status string, limit, offset int64) ([]*entity.Listing, error)
	ListByUser(ctx context.Context, userID int64, listType, status string, limit, offset int64) ([]*entity.Listing, error)
	UpdateStatusInTx(ctx context.Context, session sqlx.Session, id int64, status string) error
	RestoreAmountInTx(ctx context.Context, session sqlx.Session, id int64, amount float64) error
}

type listingRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) ListingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) Create(ctx context.Context, l *entity.Listing) (int64, error) {
	var id int64
	err := r.db.QueryRowCtx(ctx, &id,
		`INSERT INTO listings (user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, platform_fee_base, platform_fee_rate, payment_fee_base, payment_fee_rate,
		  payment_time_limit, payment_method_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		 RETURNING id`,
		l.UserID, l.Type, l.CryptoCurrency, l.FiatCurrency, l.TotalAmount, l.RemainingAmount, l.Price,
		l.MinOrderFiat, l.MaxOrderFiat, l.PlatformFeeBase, l.PlatformFeeRate, l.PaymentFeeBase, l.PaymentFeeRate,
		l.PaymentTimeLimit, l.PaymentMethodID, l.Status,
	)
	return id, err
}

func (r *listingRepository) FindByID(ctx context.Context, id int64) (*entity.Listing, error) {
	var l entity.Listing
	err := r.db.QueryRowCtx(ctx, &l,
		`SELECT `+listingSelectCols+` FROM listings WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) List(ctx context.Context, listType, status string, limit, offset int64) ([]*entity.Listing, error) {
	query := `SELECT ` + listingSelectCols + ` FROM listings WHERE 1=1`
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

	var rows []*entity.Listing
	err := r.db.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (r *listingRepository) ListByUser(ctx context.Context, userID int64, listType, status string, limit, offset int64) ([]*entity.Listing, error) {
	query := `SELECT ` + listingSelectCols + ` FROM listings WHERE user_id = $1`
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

	var rows []*entity.Listing
	err := r.db.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (r *listingRepository) UpdateStatusInTx(ctx context.Context, session sqlx.Session, id int64, status string) error {
	_, err := session.ExecCtx(ctx,
		`UPDATE listings SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *listingRepository) RestoreAmountInTx(ctx context.Context, session sqlx.Session, id int64, amount float64) error {
	_, err := session.ExecCtx(ctx,
		`UPDATE listings SET remaining_amount = remaining_amount + $1, updated_at = NOW() WHERE id = $2`,
		amount, id,
	)
	return err
}
