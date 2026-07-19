package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CryptoDeposit struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Currency    string     `db:"currency"`
	Amount      string     `db:"amount"`
	TxHash      string     `db:"tx_hash"`
	FromAddress string     `db:"from_address"`
	Memo        *string    `db:"memo"`
	Status      string     `db:"status"`
	ConfirmedAt *time.Time `db:"confirmed_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type CryptoDepositModel struct{ conn sqlx.SqlConn }

func NewCryptoDepositModel(conn sqlx.SqlConn) *CryptoDepositModel {
	return &CryptoDepositModel{conn: conn}
}

func (m *CryptoDepositModel) FindByTxHash(ctx context.Context, txHash string) (*CryptoDeposit, error) {
	var d CryptoDeposit
	err := m.conn.QueryRowCtx(ctx, &d,
		`SELECT id, user_id, currency, amount::text, tx_hash, from_address, memo, status, confirmed_at, created_at, updated_at
		 FROM crypto_deposits WHERE tx_hash = $1`,
		txHash,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (m *CryptoDepositModel) Create(ctx context.Context, d *CryptoDeposit) error {
	return m.conn.QueryRowCtx(ctx, &d.ID,
		`INSERT INTO crypto_deposits (user_id, currency, amount, tx_hash, from_address, memo, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		d.UserID, d.Currency, d.Amount, d.TxHash, d.FromAddress, d.Memo, d.Status,
	)
}

func (m *CryptoDepositModel) UpdateConfirmed(ctx context.Context, id int64, confirmedAt time.Time) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE crypto_deposits SET status='confirmed', confirmed_at=$2, updated_at=NOW() WHERE id=$1`,
		id, confirmedAt,
	)
	return err
}

func (m *CryptoDepositModel) ListPending(ctx context.Context) ([]*CryptoDeposit, error) {
	var rows []*CryptoDeposit
	err := m.conn.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, amount::text, tx_hash, from_address, memo, status, confirmed_at, created_at, updated_at
		 FROM crypto_deposits WHERE status='pending' ORDER BY created_at ASC LIMIT 100`,
	)
	return rows, err
}
