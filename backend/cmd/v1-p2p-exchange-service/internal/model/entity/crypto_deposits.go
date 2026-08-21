package entity

import "time"

// CryptoDeposit 對應 crypto_deposits 表（鏈上 USDT 充值記錄）。
// Status: pending（等待區塊確認）/ confirmed（已入帳）/ failed（memo 無法識別或確認超時）。
type CryptoDeposit struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Currency    string     `db:"currency"`
	Amount      string     `db:"amount"`
	TxHash      string     `db:"tx_hash"`
	FromAddress string     `db:"from_address"`
	Memo        *string    `db:"memo"`
	Status      string     `db:"status"`
	ConfirmedAt *time.Time `db:"confirmed_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
