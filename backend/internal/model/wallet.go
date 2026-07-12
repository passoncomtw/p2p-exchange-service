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

func (m *WalletModel) Deposit(ctx context.Context, userID int64, currency, amount string) (*Wallet, error) {
	var wallet Wallet
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		err := session.QueryRowCtx(ctx, &wallet,
			`INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
			 VALUES ($1, $2, $3, 0)
			 ON CONFLICT (user_id, currency) DO UPDATE
			 SET available_balance = wallets.available_balance + $3,
			     updated_at = NOW()
			 RETURNING id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at`,
			userID, currency, amount,
		)
		if err != nil {
			return err
		}
		_, err = session.ExecCtx(ctx,
			`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after)
			 VALUES ($1, 'deposit', $2, $3)`,
			wallet.ID, amount, wallet.AvailableBalance,
		)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (m *WalletModel) Create(ctx context.Context, userID int64, currency string) error {
	_, err := m.conn.ExecCtx(ctx,
		`INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
		 VALUES ($1, $2, 0, 0)`,
		userID, currency,
	)
	return err
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
