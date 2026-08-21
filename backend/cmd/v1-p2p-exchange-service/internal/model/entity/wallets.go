package entity

import "time"

// Wallet 對應 wallets 表。
// 餘額欄位為 NUMERIC(38, 18)，以字串承接（查詢時一律 ::text）避免 float 精度誤差。
type Wallet struct {
	ID               int64     `db:"id"`
	UserID           int64     `db:"user_id"`
	Currency         string    `db:"currency"`
	AvailableBalance string    `db:"available_balance"`
	FrozenBalance    string    `db:"frozen_balance"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// WalletLedger 對應 wallet_ledgers 表（append-only 帳本流水）。
// Amount 正數為增加、負數為減少；BalanceAfter 為異動後的餘額快照（稽核用）。
type WalletLedger struct {
	ID           int64     `db:"id"`
	WalletID     int64     `db:"wallet_id"`
	Type         string    `db:"type"`
	Amount       string    `db:"amount"`
	BalanceAfter string    `db:"balance_after"`
	RefOrderNo   *string   `db:"ref_order_no"`
	CreatedAt    time.Time `db:"created_at"`
}
