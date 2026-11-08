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
	// FindByIDForUpdateInTx 在呼叫端傳入的交易 session 內以 SELECT ... FOR UPDATE 讀取並鎖住掛單列，
	// 直到交易結束。建單、釋放訂單配置與取消掛單都必須先取得這把列鎖，彼此才會序列化。
	FindByIDForUpdateInTx(ctx context.Context, session sqlx.Session, id int64) (*entity.Listing, error)
	// DeductAmountInTx 在交易內以條件式 UPDATE 扣減掛單剩餘量，只在掛單仍為 active
	// 且剩餘量足夠時生效，回傳影響列數；影響 0 列代表不可扣減，呼叫端必須回滾整筆交易。
	DeductAmountInTx(ctx context.Context, session sqlx.Session, id int64, amount float64) (int64, error)
	// CancelInTx 在交易內以條件式 UPDATE 把 active 或 paused 的掛單轉為 cancelled，回傳影響列數；
	// 影響 0 列代表掛單狀態已被其他請求改變，呼叫端不得再解凍任何額度。
	CancelInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error)
	RestoreAmountInTx(ctx context.Context, session sqlx.Session, id int64, amount float64) error
}

const (
	findByIDForUpdateQuery = `SELECT ` + listingSelectCols + ` FROM listings WHERE id = $1 FOR UPDATE`

	deductAmountQuery = `UPDATE listings SET remaining_amount = remaining_amount - $1, updated_at = NOW()
		 WHERE id = $2 AND status = 'active' AND remaining_amount >= $1`

	cancelQuery = `UPDATE listings SET status = 'cancelled', updated_at = NOW()
		 WHERE id = $1 AND status IN ('active', 'paused')`
)

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

func (r *listingRepository) FindByIDForUpdateInTx(ctx context.Context, session sqlx.Session, id int64) (*entity.Listing, error) {
	var l entity.Listing
	if err := session.QueryRowCtx(ctx, &l, findByIDForUpdateQuery, id); err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) DeductAmountInTx(ctx context.Context, session sqlx.Session, id int64, amount float64) (int64, error) {
	return execRowsAffected(ctx, session, deductAmountQuery, amount, id)
}

func (r *listingRepository) CancelInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error) {
	return execRowsAffected(ctx, session, cancelQuery, id)
}

func (r *listingRepository) RestoreAmountInTx(ctx context.Context, session sqlx.Session, id int64, amount float64) error {
	_, err := session.ExecCtx(ctx,
		`UPDATE listings SET remaining_amount = remaining_amount + $1, updated_at = NOW() WHERE id = $2`,
		amount, id,
	)
	return err
}

// execRowsAffected 在交易 session 內執行 UPDATE 並回傳影響列數，讓呼叫端以列數判斷條件是否成立。
func execRowsAffected(ctx context.Context, session sqlx.Session, query string, args ...interface{}) (int64, error) {
	result, err := session.ExecCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
