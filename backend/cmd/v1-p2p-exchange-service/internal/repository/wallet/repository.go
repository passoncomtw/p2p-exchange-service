package walletrepo

import (
	"context"
	"fmt"
	"math/big"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/infra/rdb"
)

const walletLockTTL = 5

type walletRow struct {
	ID               int64  `db:"id"`
	UserID           int64  `db:"user_id"`
	Currency         string `db:"currency"`
	AvailableBalance string `db:"available_balance"`
	FrozenBalance    string `db:"frozen_balance"`
}

type WalletRepository interface {
	Freeze(ctx context.Context, userID int64, currency string, amount float64) error
	UnfreezeInTx(ctx context.Context, session sqlx.Session, userID int64, currency string, amount float64) error
}

type walletRepository struct {
	db  sqlx.SqlConn
	rdb *rdb.Client
}

func New(db sqlx.SqlConn, rdb *rdb.Client) WalletRepository {
	return &walletRepository{db: db, rdb: rdb}
}

func walletLockKey(userID int64, currency string) string {
	return fmt.Sprintf("lock:wallet:%d:%s", userID, currency)
}

func (r *walletRepository) Freeze(ctx context.Context, userID int64, currency string, amount float64) error {
	if r.rdb != nil {
		unlock, err := r.rdb.AcquireLock(ctx, walletLockKey(userID, currency), walletLockTTL*1e9)
		if err != nil {
			return err
		}
		defer unlock()
	}

	amountStr := fmt.Sprintf("%.18f", amount)
	return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var w walletRow
		if err := session.QueryRowCtx(ctx, &w,
			`SELECT id, user_id, currency, available_balance::text, frozen_balance::text
			 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
			userID, currency,
		); err != nil {
			if err == sqlx.ErrNotFound {
				return apierrors.New(400, "wallet not found for currency "+currency)
			}
			return err
		}

		avail, _, _ := new(big.Float).Parse(w.AvailableBalance, 10)
		if avail.Cmp(new(big.Float).SetFloat64(amount)) < 0 {
			return apierrors.New(400, "insufficient balance")
		}

		var newAvail string
		if err := session.QueryRowCtx(ctx, &newAvail,
			`UPDATE wallets SET available_balance = available_balance - $2, frozen_balance = frozen_balance + $2, updated_at = NOW()
			 WHERE id = $1 RETURNING available_balance::text`,
			w.ID, amountStr,
		); err != nil {
			return err
		}

		_, err := session.ExecCtx(ctx,
			`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after) VALUES ($1, 'freeze', -$2::numeric, $3)`,
			w.ID, amountStr, newAvail,
		)
		return err
	})
}

func (r *walletRepository) UnfreezeInTx(ctx context.Context, session sqlx.Session, userID int64, currency string, amount float64) error {
	amountStr := fmt.Sprintf("%.18f", amount)

	var w walletRow
	if err := session.QueryRowCtx(ctx, &w,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text
		 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		userID, currency,
	); err != nil {
		return err
	}

	var newAvail string
	if err := session.QueryRowCtx(ctx, &newAvail,
		`UPDATE wallets SET frozen_balance = frozen_balance - $2, available_balance = available_balance + $2, updated_at = NOW()
		 WHERE id = $1 RETURNING available_balance::text`,
		w.ID, amountStr,
	); err != nil {
		return err
	}

	_, err := session.ExecCtx(ctx,
		`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after, ref_order_no) VALUES ($1, 'unfreeze', $2, $3, '')`,
		w.ID, amountStr, newAvail,
	)
	return err
}
