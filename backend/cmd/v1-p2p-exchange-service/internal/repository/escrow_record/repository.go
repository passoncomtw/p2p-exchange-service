package escrowrepo

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type EscrowRecord struct {
	OrderID        int64
	CryptoCurrency string
	Amount         decimal.Decimal
	Action         string
	Status         string
}

const insertEscrowQuery = `INSERT INTO escrow_records (order_id, crypto_currency, amount, action, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`

type EscrowRepository interface {
	Create(ctx context.Context, r *EscrowRecord) error
	// CreateInTx 與 Create 相同，但在呼叫端傳入的交易 session 內寫入，與建立訂單屬於同一個原子交易。
	CreateInTx(ctx context.Context, session sqlx.Session, r *EscrowRecord) error
}

type escrowRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) EscrowRepository {
	return &escrowRepository{db: db}
}

func (r *escrowRepository) Create(ctx context.Context, rec *EscrowRecord) error {
	var id int64
	return r.db.QueryRowCtx(ctx, &id, insertEscrowQuery, insertEscrowArgs(rec)...)
}

func (r *escrowRepository) CreateInTx(ctx context.Context, session sqlx.Session, rec *EscrowRecord) error {
	var id int64
	return session.QueryRowCtx(ctx, &id, insertEscrowQuery, insertEscrowArgs(rec)...)
}

func insertEscrowArgs(rec *EscrowRecord) []interface{} {
	return []interface{}{rec.OrderID, rec.CryptoCurrency, rec.Amount, rec.Action, rec.Status}
}
