package model

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apperrors "p2p-exchange/internal/errors"
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

// Freeze atomically deducts amount from available_balance and adds to frozen_balance.
// Returns 400 if wallet not found or balance insufficient.
func (m *WalletModel) Freeze(ctx context.Context, userID int64, currency string, amount float64) error {
	amountStr := fmt.Sprintf("%.18f", amount)
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var wallet Wallet
		if err := session.QueryRowCtx(ctx, &wallet,
			`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
			 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
			userID, currency,
		); err != nil {
			if err == sqlx.ErrNotFound {
				return apperrors.New(400, "wallet not found for currency "+currency)
			}
			return err
		}

		avail, _, _ := new(big.Float).Parse(wallet.AvailableBalance, 10)
		if avail.Cmp(new(big.Float).SetFloat64(amount)) < 0 {
			return apperrors.New(400, "insufficient balance")
		}

		var newAvail string
		if err := session.QueryRowCtx(ctx, &newAvail,
			`UPDATE wallets SET available_balance = available_balance - $2, frozen_balance = frozen_balance + $2, updated_at = NOW()
			 WHERE id = $1 RETURNING available_balance::text`,
			wallet.ID, amountStr,
		); err != nil {
			return err
		}

		_, err := session.ExecCtx(ctx,
			`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after) VALUES ($1, 'freeze', -$2::numeric, $3)`,
			wallet.ID, amountStr, newAvail,
		)
		return err
	})
}

// UnfreezeInTx deducts from frozen_balance and adds to available_balance within a caller-provided session.
func (m *WalletModel) UnfreezeInTx(ctx context.Context, session sqlx.Session, sellerID int64, currency string, amount float64, orderNo string) error {
	amountStr := fmt.Sprintf("%.18f", amount)

	var wallet Wallet
	if err := session.QueryRowCtx(ctx, &wallet,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
		 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		sellerID, currency,
	); err != nil {
		return err
	}

	var newAvail string
	if err := session.QueryRowCtx(ctx, &newAvail,
		`UPDATE wallets SET frozen_balance = frozen_balance - $2, available_balance = available_balance + $2, updated_at = NOW()
		 WHERE id = $1 RETURNING available_balance::text`,
		wallet.ID, amountStr,
	); err != nil {
		return err
	}

	_, err := session.ExecCtx(ctx,
		`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after, ref_order_no) VALUES ($1, 'unfreeze', $2, $3, $4)`,
		wallet.ID, amountStr, newAvail, orderNo,
	)
	return err
}

// TransferInTx moves amount from seller's frozen_balance to buyer's available_balance within a caller-provided session.
func (m *WalletModel) TransferInTx(ctx context.Context, session sqlx.Session, sellerID, buyerID int64, currency string, amount float64, orderNo string) error {
	amountStr := fmt.Sprintf("%.18f", amount)
	negAmountStr := fmt.Sprintf("-%.18f", amount)

	// Seller: deduct from frozen
	var sellerWallet Wallet
	if err := session.QueryRowCtx(ctx, &sellerWallet,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
		 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		sellerID, currency,
	); err != nil {
		return err
	}

	var sellerNewFrozen string
	if err := session.QueryRowCtx(ctx, &sellerNewFrozen,
		`UPDATE wallets SET frozen_balance = frozen_balance - $2, updated_at = NOW()
		 WHERE id = $1 RETURNING frozen_balance::text`,
		sellerWallet.ID, amountStr,
	); err != nil {
		return err
	}

	if _, err := session.ExecCtx(ctx,
		`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after, ref_order_no) VALUES ($1, 'transfer_out', $2, $3, $4)`,
		sellerWallet.ID, negAmountStr, sellerNewFrozen, orderNo,
	); err != nil {
		return err
	}

	// Buyer: add to available (upsert wallet if not exists)
	var buyerWallet Wallet
	if err := session.QueryRowCtx(ctx, &buyerWallet,
		`INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
		 VALUES ($1, $2, $3, 0)
		 ON CONFLICT (user_id, currency) DO UPDATE
		 SET available_balance = wallets.available_balance + $3, updated_at = NOW()
		 RETURNING id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at`,
		buyerID, currency, amountStr,
	); err != nil {
		return err
	}

	_, err := session.ExecCtx(ctx,
		`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after, ref_order_no) VALUES ($1, 'transfer_in', $2, $3, $4)`,
		buyerWallet.ID, amountStr, buyerWallet.AvailableBalance, orderNo,
	)
	return err
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
