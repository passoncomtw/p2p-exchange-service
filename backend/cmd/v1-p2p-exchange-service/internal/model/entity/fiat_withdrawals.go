package entity

import "time"

// FiatWithdrawal 對應 fiat_withdrawals 表（TWD 提領申請記錄）。
// Status: pending（等待後台審核）/ approved（核可並扣款）/ rejected（拒絕，餘額解凍）。
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
