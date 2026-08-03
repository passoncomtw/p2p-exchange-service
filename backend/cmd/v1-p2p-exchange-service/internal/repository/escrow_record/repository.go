package escrowrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type EscrowRecord struct {
	OrderID        int64
	CryptoCurrency string
	Amount         float64
	Action         string
	Status         string
}

type EscrowRepository interface {
	Create(ctx context.Context, r *EscrowRecord) error
}

type escrowRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) EscrowRepository {
	return &escrowRepository{db: db}
}

func (r *escrowRepository) Create(ctx context.Context, rec *EscrowRecord) error {
	var id int64
	return r.db.QueryRowCtx(ctx, &id,
		`INSERT INTO escrow_records (order_id, crypto_currency, amount, action, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`,
		rec.OrderID, rec.CryptoCurrency, rec.Amount, rec.Action, rec.Status,
	)
}
