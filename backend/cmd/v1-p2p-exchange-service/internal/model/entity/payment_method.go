package entity

import "time"

type PaymentMethod struct {
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	Type          string    `db:"type"`
	BankName      string    `db:"bank_name"`
	AccountName   string    `db:"account_name"`
	AccountNumber string    `db:"account_number"`
	IsActive      bool      `db:"is_active"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
