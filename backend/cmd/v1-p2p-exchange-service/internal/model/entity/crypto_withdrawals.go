package entity

import "time"

// CryptoWithdrawal 對應 crypto_withdrawals 表（鏈上 USDT 提領記錄）。
// Status: pending（等待廣播）/ broadcasting（已廣播）/ confirmed（已確認扣款）/ failed（廣播失敗，餘額解凍）。
type CryptoWithdrawal struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Currency    string     `db:"currency"`
	Amount      string     `db:"amount"`
	ToAddress   string     `db:"to_address"`
	TxHash      *string    `db:"tx_hash"`
	Status      string     `db:"status"`
	BroadcastAt *time.Time `db:"broadcast_at"`
	ConfirmedAt *time.Time `db:"confirmed_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
