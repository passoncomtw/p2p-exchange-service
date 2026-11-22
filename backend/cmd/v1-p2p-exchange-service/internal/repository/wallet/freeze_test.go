package walletrepo

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apierrors "p2p-exchange/internal/errors"
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

// noDBWalletRepository 的 db 內嵌 nil 的 sqlx.SqlConn、rdb 為 nil：只要開交易或送出任何 SQL 就會 panic，
// 用來證明 Freeze 在金額驗證失敗時完全不碰 Redis 與資料庫。
func noDBWalletRepository() *walletRepository {
	return &walletRepository{db: noSQLConn{}}
}

type noSQLConn struct {
	sqlx.SqlConn
}

// TestFreezeRejectsNonPositiveAmount 迴歸測試：decimal 版 Freeze 必須比照 FreezeInTx 拒絕非正數。
// 掛單建立時直接以請求的 totalAmount 呼叫 Freeze；若少了驗證，負數金額會讓
// `available_balance - $amount` 反向增加可用餘額，等同憑空造幣。
// 超過 18 位小數的極小值會被四捨五入成 0 或改變數值，同樣必須被拒絕。
func TestFreezeRejectsNonPositiveAmount(t *testing.T) {
	r := noDBWalletRepository()
	ctx := context.Background()

	for _, input := range []string{"-1", "-0.000001", "0", "-0", "0.0000000000000000001", "123456789012345678901"} {
		amount := decimal.RequireFromString(input)
		err := r.Freeze(ctx, 1, "USDT", amount)
		if err == nil {
			t.Errorf("Freeze(amount=%s) = nil, want error", input)
			continue
		}
		var appErr *apierrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != 400 {
			t.Errorf("Freeze(amount=%s) = %v, want 400 AppError", input, err)
		}
	}
}
