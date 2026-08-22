package cryptodepositrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// ErrDuplicateTxHash 代表該 tx_hash 已存在（違反 crypto_deposits 的 UNIQUE 約束）。
// 交易內的 INSERT 一旦違反約束，整個交易已被 PostgreSQL 標記為 abort，
// 無法在同一交易內改查既有記錄，因此 CreateInTx 只能把這個錯誤往上拋，
// 由呼叫端判斷為「同一筆鏈上交易已被處理」並放棄本次入帳。
var ErrDuplicateTxHash = errors.New("crypto deposit tx_hash already exists")

// uniqueViolationCode PostgreSQL unique_violation 的 SQLSTATE。
const uniqueViolationCode = "23505"

// defaultPendingLimit ListPending 未指定（或給了非正數）limit 時的預設上限。
const defaultPendingLimit = 100

// cryptoDepositColumns 查詢共用欄位；amount 以 ::text 取出避免浮點誤差。
const cryptoDepositColumns = `id, user_id, currency, amount::text, tx_hash, from_address, memo, status, confirmed_at, created_at, updated_at`

// CryptoDepositRepository 存取 crypto_deposits 表（鏈上 USDT 充值記錄）。
type CryptoDepositRepository interface {
	// Create 建立一筆充值記錄並回傳寫入後的完整記錄。
	// tx_hash 已存在時視為冪等：回傳既有記錄與 nil error（不會重複建立）。
	Create(ctx context.Context, d *entity.CryptoDeposit) (*entity.CryptoDeposit, error)
	// CreateInTx 在呼叫端傳入的交易 session 內建立記錄，供「建立即已確認、需同時入帳」的路徑使用。
	// tx_hash 重複時回傳包裝 ErrDuplicateTxHash 的錯誤，交易須整筆回滾（不可在同一交易內續作查詢）。
	CreateInTx(ctx context.Context, session sqlx.Session, d *entity.CryptoDeposit) (*entity.CryptoDeposit, error)
	// FindByTxHash 依鏈上交易雜湊查詢；查無資料時回傳 sqlx.ErrNotFound。
	FindByTxHash(ctx context.Context, txHash string) (*entity.CryptoDeposit, error)
	// ListPending 取出待確認的充值記錄（建立時間由舊到新），limit <= 0 時使用預設上限。
	ListPending(ctx context.Context, limit int) ([]*entity.CryptoDeposit, error)
	// ListByUserID 分頁查詢使用者的充值記錄，依建立時間新到舊排序。
	ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.CryptoDeposit, error)
	// CountByUserID 回傳使用者充值記錄總筆數。
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	// ConfirmInTx 在呼叫端傳入的交易 session 內把狀態由 pending 轉為 confirmed，回傳異動筆數。
	// 回傳 0 代表該筆已被其他流程確認過，呼叫端必須視為冪等成功並「不再入帳」，
	// 否則同一筆充值會被重複計入餘額。
	ConfirmInTx(ctx context.Context, session sqlx.Session, id int64, confirmedAt time.Time) (int64, error)
}

type cryptoDepositRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) CryptoDepositRepository {
	return &cryptoDepositRepository{db: db}
}

// isUniqueViolation 判斷錯誤是否為 PostgreSQL 的 unique_violation。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

const insertCryptoDepositSQL = `INSERT INTO crypto_deposits (user_id, currency, amount, tx_hash, from_address, memo, status, confirmed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING ` + cryptoDepositColumns

func (r *cryptoDepositRepository) Create(ctx context.Context, d *entity.CryptoDeposit) (*entity.CryptoDeposit, error) {
	var created entity.CryptoDeposit
	err := r.db.QueryRowCtx(ctx, &created, insertCryptoDepositSQL,
		d.UserID, d.Currency, d.Amount, d.TxHash, d.FromAddress, d.Memo, d.Status, d.ConfirmedAt,
	)
	if err == nil {
		return &created, nil
	}
	// tx_hash 已存在：同一筆鏈上交易被掃描兩次，回傳既有記錄視為冪等成功。
	// 這是 DB 層的最後一道防線，主要去重仍靠呼叫端寫入前的 FindByTxHash。
	if isUniqueViolation(err) {
		return r.FindByTxHash(ctx, d.TxHash)
	}
	return nil, err
}

func (r *cryptoDepositRepository) CreateInTx(ctx context.Context, session sqlx.Session, d *entity.CryptoDeposit) (*entity.CryptoDeposit, error) {
	var created entity.CryptoDeposit
	err := session.QueryRowCtx(ctx, &created, insertCryptoDepositSQL,
		d.UserID, d.Currency, d.Amount, d.TxHash, d.FromAddress, d.Memo, d.Status, d.ConfirmedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateTxHash, d.TxHash)
		}
		return nil, err
	}
	return &created, nil
}

func (r *cryptoDepositRepository) FindByTxHash(ctx context.Context, txHash string) (*entity.CryptoDeposit, error) {
	var d entity.CryptoDeposit
	err := r.db.QueryRowCtx(ctx, &d,
		`SELECT `+cryptoDepositColumns+` FROM crypto_deposits WHERE tx_hash = $1`,
		txHash,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *cryptoDepositRepository) ListPending(ctx context.Context, limit int) ([]*entity.CryptoDeposit, error) {
	if limit <= 0 {
		limit = defaultPendingLimit
	}
	var rows []*entity.CryptoDeposit
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+cryptoDepositColumns+` FROM crypto_deposits
		 WHERE status = 'pending' ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	return rows, err
}

func (r *cryptoDepositRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.CryptoDeposit, error) {
	var rows []*entity.CryptoDeposit
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+cryptoDepositColumns+` FROM crypto_deposits
		 WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	return rows, err
}

func (r *cryptoDepositRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM crypto_deposits WHERE user_id = $1`,
		userID,
	)
	return count, err
}

func (r *cryptoDepositRepository) ConfirmInTx(ctx context.Context, session sqlx.Session, id int64, confirmedAt time.Time) (int64, error) {
	// 條件式 UPDATE：WHERE 帶 status = 'pending'，重複觸發時不會異動任何列，
	// 由 RowsAffected 讓呼叫端判斷是否真的完成了 pending → confirmed 的狀態轉換。
	// legacy 版本沒有這道守衛，重跑確認流程會重複入帳。
	res, err := session.ExecCtx(ctx,
		`UPDATE crypto_deposits SET status = 'confirmed', confirmed_at = $2, updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'`,
		id, confirmedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
