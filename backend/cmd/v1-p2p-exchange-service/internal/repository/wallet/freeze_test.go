package walletrepo

import (
	"context"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// noSQLSession 內嵌 nil 的 sqlx.Session：只要有任何 SQL 被送出就會 panic，
// 用來證明「金額驗證失敗時不會碰到資料庫」。
type noSQLSession struct {
	sqlx.Session
}

// TestFreezeInTxRejectsInvalidAmount 迴歸測試：FreezeInTx 必須在送出任何 SQL 之前擋下非法金額。
// 若少了 validateAmount，負數金額會讓 `available_balance - $amount` 反向增加可用餘額，
// 同時把負數寫進 frozen_balance，等同無中生有的提領額度。
func TestFreezeInTxRejectsInvalidAmount(t *testing.T) {
	r := &walletRepository{}
	ctx := context.Background()

	// 負數、零、正號、科學記號等一律在進入交易前被拒絕（測試通過的前提是完全沒有 SQL 被執行）。
	for _, amount := range []string{"", "0", "-1", "-0.01", "+100", "1e5", "abc"} {
		if err := r.FreezeInTx(ctx, noSQLSession{}, 1, "TWD", amount); err == nil {
			t.Errorf("FreezeInTx(amount=%q) = nil, want error", amount)
		}
	}
}
