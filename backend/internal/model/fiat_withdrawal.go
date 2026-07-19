package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FiatWithdrawal struct {
	ID           int64      `db:"id"`
	UserID       int64      `db:"user_id"`
	Currency     string     `db:"currency"`
	Amount       string     `db:"amount"`
	BankCode     string     `db:"bank_code"`
	BankAccount  string     `db:"bank_account"`
	AccountName  string     `db:"account_name"`
	Status       string     `db:"status"`
	ReviewedBy   *int64     `db:"reviewed_by"`
	ReviewedAt   *time.Time `db:"reviewed_at"`
	RejectReason *string    `db:"reject_reason"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type FiatWithdrawalModel struct{ conn sqlx.SqlConn }

func NewFiatWithdrawalModel(conn sqlx.SqlConn) *FiatWithdrawalModel {
	return &FiatWithdrawalModel{conn: conn}
}

func (m *FiatWithdrawalModel) Create(ctx context.Context, w *FiatWithdrawal) error {
	return m.conn.QueryRowCtx(ctx, &w.ID,
		`INSERT INTO fiat_withdrawals (user_id, currency, amount, bank_code, bank_account, account_name, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending') RETURNING id`,
		w.UserID, w.Currency, w.Amount, w.BankCode, w.BankAccount, w.AccountName,
	)
}

func (m *FiatWithdrawalModel) FindByID(ctx context.Context, id int64) (*FiatWithdrawal, error) {
	var w FiatWithdrawal
	err := m.conn.QueryRowCtx(ctx, &w,
		`SELECT id, user_id, currency, amount::text, bank_code, bank_account, account_name, status, reviewed_by, reviewed_at, reject_reason, created_at, updated_at
		 FROM fiat_withdrawals WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *FiatWithdrawalModel) ListByStatus(ctx context.Context, status string, limit, offset int64) ([]*FiatWithdrawal, error) {
	var rows []*FiatWithdrawal
	var err error
	if status == "" || status == "all" {
		err = m.conn.QueryRowsCtx(ctx, &rows,
			`SELECT id, user_id, currency, amount::text, bank_code, bank_account, account_name, status, reviewed_by, reviewed_at, reject_reason, created_at, updated_at
			 FROM fiat_withdrawals ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	} else {
		err = m.conn.QueryRowsCtx(ctx, &rows,
			`SELECT id, user_id, currency, amount::text, bank_code, bank_account, account_name, status, reviewed_by, reviewed_at, reject_reason, created_at, updated_at
			 FROM fiat_withdrawals WHERE status=$1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
			status, limit, offset,
		)
	}
	return rows, err
}

func (m *FiatWithdrawalModel) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	var err error
	if status == "" || status == "all" {
		err = m.conn.QueryRowCtx(ctx, &count, `SELECT COUNT(*) FROM fiat_withdrawals`)
	} else {
		err = m.conn.QueryRowCtx(ctx, &count, `SELECT COUNT(*) FROM fiat_withdrawals WHERE status=$1`, status)
	}
	return count, err
}

func (m *FiatWithdrawalModel) UpdateApproved(ctx context.Context, id, reviewedBy int64, reviewedAt time.Time) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE fiat_withdrawals SET status='approved', reviewed_by=$2, reviewed_at=$3, updated_at=NOW() WHERE id=$1`,
		id, reviewedBy, reviewedAt,
	)
	return err
}

func (m *FiatWithdrawalModel) UpdateRejected(ctx context.Context, id, reviewedBy int64, reviewedAt time.Time, reason string) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE fiat_withdrawals SET status='rejected', reviewed_by=$2, reviewed_at=$3, reject_reason=$4, updated_at=NOW() WHERE id=$1`,
		id, reviewedBy, reviewedAt, reason,
	)
	return err
}
