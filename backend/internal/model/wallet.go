package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Wallet struct {
	ID               int64     `db:"id"`
	UserID           int64     `db:"user_id"`
	Currency         string    `db:"currency"`
	AvailableBalance string    `db:"available_balance"`
	FrozenBalance    string    `db:"frozen_balance"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type WalletModel struct {
	conn sqlx.SqlConn
}

func NewWalletModel(conn sqlx.SqlConn) *WalletModel {
	return &WalletModel{conn: conn}
}

func (m *WalletModel) FindOne(ctx context.Context, userID int64, currency string) (*Wallet, error) {
	var w Wallet
	err := m.conn.QueryRowCtx(ctx, &w,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
		 FROM wallets WHERE user_id = $1 AND currency = $2`,
		userID, currency,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *WalletModel) FindByUserID(ctx context.Context, userID int64) ([]*Wallet, error) {
	var rows []*Wallet
	err := m.conn.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
		 FROM wallets WHERE user_id = $1 ORDER BY currency ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
