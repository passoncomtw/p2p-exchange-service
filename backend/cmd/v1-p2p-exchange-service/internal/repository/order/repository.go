package orderrepo

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

const orderSelectCols = `id, order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
  total_fee, total_amount, payment_method_id, status, payment_deadline, paid_at, confirmed_at, completed_at,
  cancelled_at, cancel_reason, created_at, updated_at`

type OrderRepository interface {
	Create(ctx context.Context, o *entity.Order) (int64, error)
	FindByID(ctx context.Context, id int64) (*entity.Order, error)
	FindExpired(ctx context.Context) ([]*entity.Order, error)
	List(ctx context.Context, userID int64, role, status string, limit, offset int64) ([]*entity.Order, error)
	BackendList(ctx context.Context, status string, limit, offset int64) ([]*entity.Order, error)
	UpdateStatusInTx(ctx context.Context, session sqlx.Session, id int64, status string, extras map[string]interface{}) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status string, extras map[string]interface{}) error
	DeductListingAmount(ctx context.Context, listingID int64, amount float64) error
}

type orderRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, o *entity.Order) (int64, error) {
	var id int64
	err := r.db.QueryRowCtx(ctx, &id,
		`INSERT INTO orders (order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
		  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
		  total_fee, total_amount, payment_method_id, status, payment_deadline, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW(), NOW())
		 RETURNING id`,
		o.OrderNo, o.ListingID, o.ListingType, o.SellerID, o.BuyerID, o.CryptoCurrency, o.FiatCurrency,
		o.CryptoAmount, o.Price, o.FiatAmount, o.PlatformFeeBase, o.PlatformFeeAmount, o.PaymentFeeBase, o.PaymentFeeAmount,
		o.TotalFee, o.TotalAmount, o.PaymentMethodID, o.Status, o.PaymentDeadline,
	)
	return id, err
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*entity.Order, error) {
	var o entity.Order
	err := r.db.QueryRowCtx(ctx, &o,
		`SELECT `+orderSelectCols+` FROM orders WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) List(ctx context.Context, userID int64, role, status string, limit, offset int64) ([]*entity.Order, error) {
	query := `SELECT ` + orderSelectCols + ` FROM orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	switch role {
	case "buyer":
		query += fmt.Sprintf(" AND buyer_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	case "seller":
		query += fmt.Sprintf(" AND seller_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	default:
		query += fmt.Sprintf(" AND (buyer_id = $%d OR seller_id = $%d)", argIdx, argIdx)
		args = append(args, userID, userID)
		argIdx++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*entity.Order
	err := r.db.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (r *orderRepository) BackendList(ctx context.Context, status string, limit, offset int64) ([]*entity.Order, error) {
	query := `SELECT ` + orderSelectCols + ` FROM orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*entity.Order
	err := r.db.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status string, extras map[string]interface{}) error {
	_, err := buildStatusUpdate(r.db, ctx, id, status, extras)
	return err
}

func (r *orderRepository) UpdateStatusInTx(ctx context.Context, session sqlx.Session, id int64, status string, extras map[string]interface{}) (int64, error) {
	query, args := buildStatusUpdateQuery(id, status, extras)
	result, err := session.ExecCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return n, nil
}

func (r *orderRepository) FindExpired(ctx context.Context) ([]*entity.Order, error) {
	var rows []*entity.Order
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+orderSelectCols+` FROM orders WHERE status = 'matched' AND payment_deadline < NOW()`,
	)
	return rows, err
}

func (r *orderRepository) DeductListingAmount(ctx context.Context, listingID int64, amount float64) error {
	_, err := r.db.ExecCtx(ctx,
		`UPDATE listings SET remaining_amount = remaining_amount - $1, updated_at = NOW()
		 WHERE id = $2 AND remaining_amount >= $1`,
		amount, listingID,
	)
	return err
}

func buildStatusUpdate(db sqlx.SqlConn, ctx context.Context, id int64, status string, extras map[string]interface{}) (int64, error) {
	query, args := buildStatusUpdateQuery(id, status, extras)
	result, err := db.ExecCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func buildStatusUpdateQuery(id int64, status string, extras map[string]interface{}) (string, []interface{}) {
	query := "UPDATE orders SET status = $1, updated_at = NOW()"
	args := []interface{}{status}
	argIdx := 2

	for key, val := range extras {
		switch key {
		case "paid_at", "confirmed_at", "completed_at", "cancelled_at", "cancel_reason":
			query += fmt.Sprintf(", %s = $%d", key, argIdx)
			args = append(args, val)
			argIdx++
		}
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)
	return query, args
}
