package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type EscrowRecord struct {
	ID             int64     `db:"id"`
	OrderID        int64     `db:"order_id"`
	CryptoCurrency string    `db:"crypto_currency"`
	Amount         float64   `db:"amount"`
	Action         string    `db:"action"`
	Status         string    `db:"status"`
	TxRef          *string   `db:"tx_ref"`
	Remark         *string   `db:"remark"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type EscrowRecordModel struct {
	conn sqlx.SqlConn
}

func NewEscrowRecordModel(conn sqlx.SqlConn) *EscrowRecordModel {
	return &EscrowRecordModel{conn: conn}
}

func (m *EscrowRecordModel) Create(ctx context.Context, r *EscrowRecord) (int64, error) {
	var id int64
	err := m.conn.QueryRowCtx(ctx, &id,
		`INSERT INTO escrow_records (order_id, crypto_currency, amount, action, status, tx_ref, remark, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id`,
		r.OrderID, r.CryptoCurrency, r.Amount, r.Action, r.Status, r.TxRef, r.Remark,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *EscrowRecordModel) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE escrow_records SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}
