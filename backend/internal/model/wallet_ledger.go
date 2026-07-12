package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type WalletLedger struct {
	ID           int64     `db:"id"`
	WalletID     int64     `db:"wallet_id"`
	Type         string    `db:"type"`
	Amount       string    `db:"amount"`
	BalanceAfter string    `db:"balance_after"`
	RefOrderNo   *string   `db:"ref_order_no"`
	CreatedAt    time.Time `db:"created_at"`
}

type WalletLedgerModel struct {
	conn sqlx.SqlConn
}

func NewWalletLedgerModel(conn sqlx.SqlConn) *WalletLedgerModel {
	return &WalletLedgerModel{conn: conn}
}

func (m *WalletLedgerModel) ListByWalletID(ctx context.Context, walletID, limit, offset int64) ([]*WalletLedger, error) {
	var rows []*WalletLedger
	err := m.conn.QueryRowsCtx(ctx, &rows,
		`SELECT id, wallet_id, type, amount::text, balance_after::text, ref_order_no, created_at
		 FROM wallet_ledgers WHERE wallet_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		walletID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *WalletLedgerModel) CountByWalletID(ctx context.Context, walletID int64) (int64, error) {
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total,
		`SELECT COUNT(*) FROM wallet_ledgers WHERE wallet_id = $1`,
		walletID,
	)
	if err != nil {
		return 0, err
	}
	return total, nil
}
