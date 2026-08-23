package fiatwithdrawrepo

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// fiatWithdrawalColumns 查詢共用欄位；amount 以 ::text 取出避免浮點誤差。
const fiatWithdrawalColumns = `id, user_id, currency, amount::text, bank_code, bank_account, account_name, status,
	        reviewed_by, reviewed_at, reject_reason, created_at, updated_at`

// statusAll ListByStatus / CountByStatus 代表「不限狀態」的兩種輸入（空字串或 "all"）。
const statusAll = "all"

// FiatWithdrawRepository 存取 fiat_withdrawals 表（TWD 提領申請記錄）。
//
// 狀態機：pending（待後台審核）→ approved（核可並自凍結餘額扣款）或 rejected（駁回並解凍）。
// 審核的狀態轉換一律用條件式 UPDATE 帶 `status = 'pending'` 守衛，並以 RowsAffected 判斷
// 是否真的由本次呼叫完成轉換。
type FiatWithdrawRepository interface {
	// CreateInTx 在呼叫端傳入的交易 session 內建立提領申請（狀態固定 'pending'），回傳寫入後的完整記錄。
	// 提領申請必須與餘額凍結在同一個交易內，避免凍結成功卻沒有對應單據（或反之）。
	CreateInTx(ctx context.Context, session sqlx.Session, w *entity.FiatWithdrawal) (*entity.FiatWithdrawal, error)
	// ListByUserID 分頁查詢使用者的提領記錄，依建立時間新到舊排序。
	ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatWithdrawal, error)
	// CountByUserID 回傳使用者提領記錄總筆數。
	CountByUserID(ctx context.Context, userID int64) (int64, error)

	// FindByID 依主鍵查詢單筆提領申請；查無資料時回傳 sqlx.ErrNotFound，
	// 由呼叫端決定要轉成 404 或其他語意。
	//
	// 注意：這是「無鎖」的純讀取，回傳的 Status 只能用於顯示或取得金額／幣別等不可變欄位，
	// 絕不可用來當作「可以審核」的判斷依據（check-then-act 非原子，legacy 的 P0 缺陷來源）。
	// 是否能推進狀態一律以 UpdateApprovedInTx / UpdateRejectedInTx 的 RowsAffected 為準。
	FindByID(ctx context.Context, id int64) (*entity.FiatWithdrawal, error)
	// ListByStatus 後台分頁查詢提領申請；status 為空字串或 "all" 時回全部（依 created_at DESC），
	// 否則只回該狀態（依 created_at ASC）。兩種排序不一致是刻意保留的 legacy 既有行為。
	ListByStatus(ctx context.Context, status string, limit, offset int64) ([]*entity.FiatWithdrawal, error)
	// CountByStatus 回傳對應 status 篩選條件的總筆數（篩選語意與 ListByStatus 一致）。
	CountByStatus(ctx context.Context, status string) (int64, error)
	// UpdateApprovedInTx 在交易 session 內把狀態由 pending 轉為 approved，回傳異動筆數。
	//
	// 回傳 0 代表該筆已被其他流程審核過（並發雙擊、前端重送），呼叫端「必須」視為衝突並
	// 中止整筆交易，不得繼續執行任何資金異動：重複核可等於自使用者凍結餘額多扣一次款。
	// 審核是人為決策，0 筆不可被當成冪等成功靜默吞掉。
	UpdateApprovedInTx(ctx context.Context, session sqlx.Session, id, reviewerID int64, reviewedAt time.Time) (int64, error)
	// UpdateRejectedInTx 在交易 session 內把狀態由 pending 轉為 rejected 並記錄駁回原因，回傳異動筆數。
	// 回傳 0 的處理方式與 UpdateApprovedInTx 相同：視為衝突，不得繼續解凍餘額，
	// 否則使用者會拿回兩倍金額。
	UpdateRejectedInTx(ctx context.Context, session sqlx.Session, id, reviewerID int64, reviewedAt time.Time, reason string) (int64, error)
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

func (r *fiatWithdrawRepository) FindByID(ctx context.Context, id int64) (*entity.FiatWithdrawal, error) {
	var row entity.FiatWithdrawal
	err := r.db.QueryRowCtx(ctx, &row,
		`SELECT `+fiatWithdrawalColumns+` FROM fiat_withdrawals WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *fiatWithdrawRepository) ListByStatus(ctx context.Context, status string, limit, offset int64) ([]*entity.FiatWithdrawal, error) {
	var rows []*entity.FiatWithdrawal
	if status == "" || status == statusAll {
		err := r.db.QueryRowsCtx(ctx, &rows,
			`SELECT `+fiatWithdrawalColumns+` FROM fiat_withdrawals
			 ORDER BY created_at DESC
			 LIMIT $1 OFFSET $2`,
			limit, offset,
		)
		return rows, err
	}

	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+fiatWithdrawalColumns+` FROM fiat_withdrawals
		 WHERE status = $1
		 ORDER BY created_at ASC
		 LIMIT $2 OFFSET $3`,
		status, limit, offset,
	)
	return rows, err
}

func (r *fiatWithdrawRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if status == "" || status == statusAll {
		err := r.db.QueryRowCtx(ctx, &count, `SELECT COUNT(*) FROM fiat_withdrawals`)
		return count, err
	}

	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM fiat_withdrawals WHERE status = $1`,
		status,
	)
	return count, err
}

func (r *fiatWithdrawRepository) UpdateApprovedInTx(ctx context.Context, session sqlx.Session, id, reviewerID int64, reviewedAt time.Time) (int64, error) {
	// 條件式 UPDATE（WHERE status = 'pending'）：同一筆申請同時被兩個後台請求審核時，
	// PostgreSQL 的列鎖讓第二個交易看到已轉為 'approved' / 'rejected' 的值，異動 0 列。
	// 呼叫端必須先取得 RowsAffected = 1 才能動用資金（legacy 缺這道守衛，會重複扣款）。
	res, err := session.ExecCtx(ctx,
		`UPDATE fiat_withdrawals
		 SET status = 'approved', reviewed_by = $2, reviewed_at = $3, updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'`,
		id, reviewerID, reviewedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *fiatWithdrawRepository) UpdateRejectedInTx(ctx context.Context, session sqlx.Session, id, reviewerID int64, reviewedAt time.Time, reason string) (int64, error) {
	res, err := session.ExecCtx(ctx,
		`UPDATE fiat_withdrawals
		 SET status = 'rejected', reviewed_by = $2, reviewed_at = $3, reject_reason = $4, updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'`,
		id, reviewerID, reviewedAt, reason,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
