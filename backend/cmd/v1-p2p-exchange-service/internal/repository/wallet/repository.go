package walletrepo

import (
	"context"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/infra/rdb"
)

const (
	walletLockTTL         = 5
	defaultLedgerPageSize = 20
)

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

	// FindByUserID 查詢使用者所有幣種的錢包，依幣別排序。
	FindByUserID(ctx context.Context, userID int64) ([]*entity.Wallet, error)
	// FindOne 查詢使用者指定幣別的錢包（純讀，不加鎖）；查無資料時回傳 sqlx.ErrNotFound，
	// 由呼叫端決定要轉成 404 或其他語意。
	FindOne(ctx context.Context, userID int64, currency string) (*entity.Wallet, error)
	// ListLedgers 分頁查詢指定錢包的帳本流水，並回傳符合條件的總筆數。
	ListLedgers(ctx context.Context, walletID int64, limit, offset int) ([]*entity.WalletLedger, int64, error)
	// Deposit 將金額入帳到 available_balance（錢包不存在則建立），並寫入 type='deposit' 帳本，
	// 回傳異動後的可用餘額與凍結餘額。
	Deposit(ctx context.Context, userID int64, currency string, amount string) (availableBalance string, frozenBalance string, err error)
	// DepositWithLedgerType 同 Deposit，但可指定帳本類型（crypto_deposit / fiat_deposit）與關聯單號。
	DepositWithLedgerType(ctx context.Context, userID int64, currency string, amount string, ledgerType string, refOrderNo string) error
	// DeductFrozenBalance 自 frozen_balance 扣除金額（提領確認），凍結餘額不足時回傳 400。
	DeductFrozenBalance(ctx context.Context, userID int64, currency string, amount string) error
	// UnfreezeBalance 將凍結金額退回 available_balance（提領失敗/駁回），凍結餘額不足時回傳 400。
	UnfreezeBalance(ctx context.Context, userID int64, currency string, amount string) error
	// SumCurrencyBalance 回傳該幣別所有使用者的 available_balance + frozen_balance 總和。
	SumCurrencyBalance(ctx context.Context, currency string) (string, error)
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

// amountPattern 只接受十進位正數金額字串（不接受正負號與科學記號）。
// 負數金額會讓 `frozen_balance >= $amount` 這類條件式檢查形同虛設（任何餘額都成立），
// 因此所有異動餘額的方法都必須先通過這道驗證。
var amountPattern = regexp.MustCompile(`^\d{1,20}(\.\d{1,18})?$`)

// validateAmount 驗證金額字串為大於零的十進位正數。
func validateAmount(amount string) error {
	if !amountPattern.MatchString(amount) || strings.Trim(amount, "0.") == "" {
		return apierrors.New(400, "invalid amount")
	}
	return nil
}

// withWalletLock 取得該使用者該幣別的分散式鎖後執行 fn（未設定 Redis 時直接執行）。
func (r *walletRepository) withWalletLock(ctx context.Context, userID int64, currency string, fn func(ctx context.Context) error) error {
	if r.rdb == nil {
		return fn(ctx)
	}
	unlock, err := r.rdb.AcquireLock(ctx, walletLockKey(userID, currency), walletLockTTL*1e9)
	if err != nil {
		return err
	}
	defer unlock()
	return fn(ctx)
}

// lockWalletRow 以 SELECT ... FOR UPDATE 鎖定並讀取錢包列；錢包不存在時回傳 400。
func lockWalletRow(ctx context.Context, session sqlx.Session, userID int64, currency string) (walletRow, error) {
	var w walletRow
	err := session.QueryRowCtx(ctx, &w,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text
		 FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		userID, currency,
	)
	if err == sqlx.ErrNotFound {
		return w, apierrors.New(400, "wallet not found for currency "+currency)
	}
	return w, err
}

func (r *walletRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.Wallet, error) {
	var rows []*entity.Wallet
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
		 FROM wallets WHERE user_id = $1 ORDER BY currency ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *walletRepository) FindOne(ctx context.Context, userID int64, currency string) (*entity.Wallet, error) {
	var w entity.Wallet
	err := r.db.QueryRowCtx(ctx, &w,
		`SELECT id, user_id, currency, available_balance::text, frozen_balance::text, created_at, updated_at
		 FROM wallets WHERE user_id = $1 AND currency = $2`,
		userID, currency,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *walletRepository) ListLedgers(ctx context.Context, walletID int64, limit, offset int) ([]*entity.WalletLedger, int64, error) {
	if limit <= 0 {
		limit = defaultLedgerPageSize
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := r.db.QueryRowCtx(ctx, &total,
		`SELECT COUNT(*) FROM wallet_ledgers WHERE wallet_id = $1`,
		walletID,
	); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*entity.WalletLedger{}, 0, nil
	}

	var rows []*entity.WalletLedger
	if err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT id, wallet_id, type, amount::text, balance_after::text, ref_order_no, created_at
		 FROM wallet_ledgers WHERE wallet_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		walletID, limit, offset,
	); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *walletRepository) Deposit(ctx context.Context, userID int64, currency string, amount string) (string, string, error) {
	if err := validateAmount(amount); err != nil {
		return "", "", err
	}

	var w walletRow
	err := r.withWalletLock(ctx, userID, currency, func(ctx context.Context) error {
		return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			if err := session.QueryRowCtx(ctx, &w,
				`INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
				 VALUES ($1, $2, $3, 0)
				 ON CONFLICT (user_id, currency) DO UPDATE
				 SET available_balance = wallets.available_balance + $3, updated_at = NOW()
				 RETURNING id, user_id, currency, available_balance::text, frozen_balance::text`,
				userID, currency, amount,
			); err != nil {
				return err
			}

			_, err := session.ExecCtx(ctx,
				`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after)
				 VALUES ($1, 'deposit', $2, $3)`,
				w.ID, amount, w.AvailableBalance,
			)
			return err
		})
	})
	if err != nil {
		return "", "", err
	}
	return w.AvailableBalance, w.FrozenBalance, nil
}

func (r *walletRepository) DepositWithLedgerType(ctx context.Context, userID int64, currency string, amount string, ledgerType string, refOrderNo string) error {
	if err := validateAmount(amount); err != nil {
		return err
	}

	return r.withWalletLock(ctx, userID, currency, func(ctx context.Context) error {
		return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			var w walletRow
			if err := session.QueryRowCtx(ctx, &w,
				`INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
				 VALUES ($1, $2, $3, 0)
				 ON CONFLICT (user_id, currency) DO UPDATE
				 SET available_balance = wallets.available_balance + $3, updated_at = NOW()
				 RETURNING id, user_id, currency, available_balance::text, frozen_balance::text`,
				userID, currency, amount,
			); err != nil {
				return err
			}

			_, err := session.ExecCtx(ctx,
				`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after, ref_order_no)
				 VALUES ($1, $2, $3, $4, $5)`,
				w.ID, ledgerType, amount, w.AvailableBalance, refOrderNo,
			)
			return err
		})
	})
}

func (r *walletRepository) DeductFrozenBalance(ctx context.Context, userID int64, currency string, amount string) error {
	if err := validateAmount(amount); err != nil {
		return err
	}

	return r.withWalletLock(ctx, userID, currency, func(ctx context.Context) error {
		return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			w, err := lockWalletRow(ctx, session, userID, currency)
			if err != nil {
				return err
			}

			// 條件式 UPDATE：凍結餘額不足時不會異動任何列，靠 RowsAffected 攔截，
			// 不可只靠先前 SELECT 的值判斷（legacy 版本沒有這道檢查，會扣成負數）。
			res, err := session.ExecCtx(ctx,
				`UPDATE wallets SET frozen_balance = frozen_balance - $2, updated_at = NOW()
				 WHERE id = $1 AND frozen_balance >= $2`,
				w.ID, amount,
			)
			if err != nil {
				return err
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return apierrors.New(400, "insufficient frozen balance")
			}

			var newFrozen string
			if err := session.QueryRowCtx(ctx, &newFrozen,
				`SELECT frozen_balance::text FROM wallets WHERE id = $1`,
				w.ID,
			); err != nil {
				return err
			}

			// balance_after 記錄異動後的凍結餘額（與既有 transfer_out 帳本一致：
			// 提領扣的是凍結餘額，available_balance 並未變動）。
			_, err = session.ExecCtx(ctx,
				`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after)
				 VALUES ($1, 'withdraw', $2, $3)`,
				w.ID, "-"+amount, newFrozen,
			)
			return err
		})
	})
}

func (r *walletRepository) UnfreezeBalance(ctx context.Context, userID int64, currency string, amount string) error {
	if err := validateAmount(amount); err != nil {
		return err
	}

	return r.withWalletLock(ctx, userID, currency, func(ctx context.Context) error {
		return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			w, err := lockWalletRow(ctx, session, userID, currency)
			if err != nil {
				return err
			}

			// 條件式 UPDATE + RowsAffected：凍結餘額不足視為衝突，回傳 400 而非靜默通過。
			res, err := session.ExecCtx(ctx,
				`UPDATE wallets SET frozen_balance = frozen_balance - $2, available_balance = available_balance + $2, updated_at = NOW()
				 WHERE id = $1 AND frozen_balance >= $2`,
				w.ID, amount,
			)
			if err != nil {
				return err
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return apierrors.New(400, "insufficient frozen balance")
			}

			var newAvail string
			if err := session.QueryRowCtx(ctx, &newAvail,
				`SELECT available_balance::text FROM wallets WHERE id = $1`,
				w.ID,
			); err != nil {
				return err
			}

			_, err = session.ExecCtx(ctx,
				`INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after)
				 VALUES ($1, 'unfreeze', $2, $3)`,
				w.ID, amount, newAvail,
			)
			return err
		})
	})
}

func (r *walletRepository) SumCurrencyBalance(ctx context.Context, currency string) (string, error) {
	var total string
	err := r.db.QueryRowCtx(ctx, &total,
		`SELECT COALESCE(SUM(available_balance + frozen_balance), 0)::text FROM wallets WHERE currency = $1`,
		currency,
	)
	return total, err
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
