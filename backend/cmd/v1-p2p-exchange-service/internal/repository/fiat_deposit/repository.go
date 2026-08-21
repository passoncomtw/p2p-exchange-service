package fiatdepositrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// FiatDepositRepository 存取 fiat_deposits 表（ECPay TWD 入金記錄）。
type FiatDepositRepository interface {
	// Create 新增一筆 pending 入金記錄，回傳新的主鍵 ID。
	Create(ctx context.Context, d *entity.FiatDeposit) (int64, error)
	// ListByUserID 分頁查詢使用者的入金記錄，依建立時間新到舊排序。
	ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatDeposit, error)
	// CountByUserID 回傳使用者入金記錄總筆數。
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}

type fiatDepositRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) FiatDepositRepository {
	return &fiatDepositRepository{db: db}
}

func (r *fiatDepositRepository) Create(ctx context.Context, d *entity.FiatDeposit) (int64, error) {
	var id int64
	err := r.db.QueryRowCtx(ctx, &id,
		`INSERT INTO fiat_deposits (user_id, currency, amount, merchant_trade_no, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		d.UserID, d.Currency, d.Amount, d.MerchantTradeNo, d.Status,
	)
	return id, err
}

func (r *fiatDepositRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatDeposit, error) {
	var rows []*entity.FiatDeposit
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, amount::text, ecpay_order_no, merchant_trade_no, status, payment_type, paid_at, created_at, updated_at
		 FROM fiat_deposits WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	return rows, err
}

func (r *fiatDepositRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM fiat_deposits WHERE user_id = $1`,
		userID,
	)
	return count, err
}
