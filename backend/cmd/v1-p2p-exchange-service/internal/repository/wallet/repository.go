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
	TransferInTx(ctx context.Context, session sqlx.Session, sellerID, buyerID int64, currency string, amount float64, orderNo string) error
	AcquireLocks(ctx context.Context, userIDs []int64, currency string) (unlock func(), err error)
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

func (r *walletRepository) TransferInTx(ctx context.Context, session sqlx.Session, sellerID, buyerID int64, currency string, amount float64, orderNo string) error {
	amountStr := fmt.Sprintf("%.18f", amount)
	negAmountStr := fmt.Sprintf("-%.18f", amount)

	var seller walletRow
	if err := session.QueryRowCtx(ctx, &seller,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text
		 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		sellerID, currency,
	); err != nil {
		return err
	}

	var sellerNewFrozen string
	if err := session.QueryRowCtx(ctx, &sellerNewFrozen,
		`UPDATE wallets SET frozen_balance = frozen_balance - $2, updated_at = NOW()
		 WHERE id = $1 RETURNING frozen_balance::text`,
		seller.ID, amountStr,
	); err != nil {
		return err
	}

	if _, err := session.ExecCtx(ctx,
		`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after, ref_order_no) VALUES ($1, 'transfer_out', $2, $3, $4)`,
		seller.ID, negAmountStr, sellerNewFrozen, orderNo,
	); err != nil {
		return err
	}

	var buyer walletRow
	if err := session.QueryRowCtx(ctx, &buyer,
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
		buyer.ID, amountStr, buyer.AvailableBalance, orderNo,
	)
	return err
}

func (r *walletRepository) AcquireLocks(ctx context.Context, userIDs []int64, currency string) (unlock func(), err error) {
	if r.rdb == nil {
		return func() {}, nil
	}
	keys := make([]string, len(userIDs))
	for i, uid := range userIDs {
		keys[i] = walletLockKey(uid, currency)
	}
	return r.rdb.AcquireLocks(ctx, keys, 5*1e9)
}
