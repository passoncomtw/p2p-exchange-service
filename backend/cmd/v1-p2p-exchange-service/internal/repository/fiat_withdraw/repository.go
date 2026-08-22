package fiatwithdrawrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// FiatWithdrawRepository 存取 fiat_withdrawals 表（TWD 提領申請記錄）。
type FiatWithdrawRepository interface {
	// CreateInTx 在呼叫端傳入的交易 session 內建立提領申請（狀態固定 'pending'），回傳寫入後的完整記錄。
	// 提領申請必須與餘額凍結在同一個交易內，避免凍結成功卻沒有對應單據（或反之）。
	CreateInTx(ctx context.Context, session sqlx.Session, w *entity.FiatWithdrawal) (*entity.FiatWithdrawal, error)
	// ListByUserID 分頁查詢使用者的提領記錄，依建立時間新到舊排序。
	ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatWithdrawal, error)
	// CountByUserID 回傳使用者提領記錄總筆數。
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}

type fiatWithdrawRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) FiatWithdrawRepository {
	return &fiatWithdrawRepository{db: db}
}

func (r *fiatWithdrawRepository) CreateInTx(ctx context.Context, session sqlx.Session, w *entity.FiatWithdrawal) (*entity.FiatWithdrawal, error) {
	// status 由 SQL 固定為 'pending'，不採用呼叫端傳入的值：
	// 審核狀態只能由後台審核流程（Slice H）推進。
	var row entity.FiatWithdrawal
	err := session.QueryRowCtx(ctx, &row,
		`INSERT INTO fiat_withdrawals (user_id, currency, amount, bank_code, bank_account, account_name, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		 RETURNING id, user_id, currency, amount::text, bank_code, bank_account, account_name, status,
		           reviewed_by, reviewed_at, reject_reason, created_at, updated_at`,
		w.UserID, w.Currency, w.Amount, w.BankCode, w.BankAccount, w.AccountName,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *fiatWithdrawRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatWithdrawal, error) {
	var rows []*entity.FiatWithdrawal
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, amount::text, bank_code, bank_account, account_name, status,
		        reviewed_by, reviewed_at, reject_reason, created_at, updated_at
		 FROM fiat_withdrawals WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	return rows, err
}

func (r *fiatWithdrawRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM fiat_withdrawals WHERE user_id = $1`,
		userID,
	)
	return count, err
}
