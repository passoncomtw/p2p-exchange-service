package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CryptoWithdrawal struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Currency    string     `db:"currency"`
	Amount      string     `db:"amount"`
	ToAddress   string     `db:"to_address"`
	TxHash      *string    `db:"tx_hash"`
	Status      string     `db:"status"`
	BroadcastAt *time.Time `db:"broadcast_at"`
	ConfirmedAt *time.Time `db:"confirmed_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type CryptoWithdrawalModel struct{ conn sqlx.SqlConn }

func NewCryptoWithdrawalModel(conn sqlx.SqlConn) *CryptoWithdrawalModel {
	return &CryptoWithdrawalModel{conn: conn}
}

func (m *CryptoWithdrawalModel) Create(ctx context.Context, w *CryptoWithdrawal) error {
	return m.conn.QueryRowCtx(ctx, &w.ID,
		`INSERT INTO crypto_withdrawals (user_id, currency, amount, to_address, status)
		 VALUES ($1, $2, $3, $4, 'pending') RETURNING id`,
		w.UserID, w.Currency, w.Amount, w.ToAddress,
	)
}

func (m *CryptoWithdrawalModel) ListPending(ctx context.Context) ([]*CryptoWithdrawal, error) {
	var rows []*CryptoWithdrawal
	err := m.conn.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, amount::text, to_address, tx_hash, status, broadcast_at, confirmed_at, created_at, updated_at
		 FROM crypto_withdrawals WHERE status='pending' ORDER BY created_at ASC LIMIT 50`,
	)
	return rows, err
}

func (m *CryptoWithdrawalModel) ListBroadcasting(ctx context.Context) ([]*CryptoWithdrawal, error) {
	var rows []*CryptoWithdrawal
	err := m.conn.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, amount::text, to_address, tx_hash, status, broadcast_at, confirmed_at, created_at, updated_at
		 FROM crypto_withdrawals WHERE status='broadcasting' ORDER BY created_at ASC LIMIT 50`,
	)
	return rows, err
}

func (m *CryptoWithdrawalModel) UpdateBroadcasting(ctx context.Context, id int64, txHash string, broadcastAt time.Time) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET status='broadcasting', tx_hash=$2, broadcast_at=$3, updated_at=NOW() WHERE id=$1`,
		id, txHash, broadcastAt,
	)
	return err
}

func (m *CryptoWithdrawalModel) UpdateConfirmed(ctx context.Context, id int64, confirmedAt time.Time) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET status='confirmed', confirmed_at=$2, updated_at=NOW() WHERE id=$1`,
		id, confirmedAt,
	)
	return err
}

func (m *CryptoWithdrawalModel) UpdateFailed(ctx context.Context, id int64) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET status='failed', updated_at=NOW() WHERE id=$1`,
		id,
	)
	return err
}
