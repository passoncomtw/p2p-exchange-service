package job

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/internal/infra/mq"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/internal/model"
)

const orderProcessLockTTL = 30 * time.Second

type ExpiredOrderDeps struct {
	RDB       *rdb.Client
	Order     *model.OrderModel
	Wallet    *model.WalletModel
	Listing   *model.ListingModel
	StatusLog *model.OrderStatusLogModel
	DB        sqlx.SqlConn
}

// StartExpiredOrderConsumer 訂閱 order.timeout.check，批次處理超時訂單。
// 每筆訂單以分散式鎖保護，避免多實例重複處理（idempotency）。
func StartExpiredOrderConsumer(mqClient *mq.Client, deps ExpiredOrderDeps) {
	if mqClient == nil {
		return
	}
	if err := mqClient.Subscribe("order.timeout.check", func(ctx context.Context, _ []byte) error {
		return processExpiredOrders(ctx, deps)
	}); err != nil {
		logx.Errorf("subscribe order.timeout.check error: %v", err)
	}
}

func processExpiredOrders(ctx context.Context, deps ExpiredOrderDeps) error {
	orders, err := deps.Order.FindExpired(ctx)
	if err != nil {
		return fmt.Errorf("find expired orders: %w", err)
	}
	for _, o := range orders {
		if err := cancelExpiredOrder(ctx, deps, o); err != nil {
			logx.Errorf("cancel expired order %d error: %v", o.ID, err)
		}
	}
	return nil
}

func cancelExpiredOrder(ctx context.Context, deps ExpiredOrderDeps, o *model.Order) error {
	// 分散式鎖確保同一訂單不被重複處理（30 秒 TTL 覆蓋整個 transaction 時間）
	if deps.RDB != nil {
		key := fmt.Sprintf("processing:order:%d", o.ID)
		unlock, err := deps.RDB.AcquireLock(ctx, key, orderProcessLockTTL)
		if err != nil {
			return nil // 已被其他實例持有，略過
		}
		defer unlock()
	}

	// 鎖定賣家錢包，防止與使用者 Freeze/Deposit 並發
	if deps.RDB != nil {
		unlock, err := deps.RDB.AcquireLock(ctx, model.WalletLockKey(o.SellerID, o.CryptoCurrency), 10*time.Second)
		if err != nil {
			return nil // 錢包正在被操作，此次略過，下個 tick 重試
		}
		defer unlock()
	}

	return deps.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. CAS 更新訂單狀態（status = 'matched' 才更新，防止重複執行）
		result, err := session.ExecCtx(ctx,
			`UPDATE orders SET status = 'timeout', cancelled_at = NOW(), updated_at = NOW()
			 WHERE id = $1 AND status = 'matched'`,
			o.ID,
		)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n == 0 {
			return nil // 已被其他流程處理
		}

		// 2. 解凍賣家餘額，寫入 wallet_ledger
		if err := deps.Wallet.UnfreezeInTx(ctx, session, o.SellerID, o.CryptoCurrency, o.CryptoAmount, o.OrderNo); err != nil {
			return fmt.Errorf("unfreeze seller wallet: %w", err)
		}

		// 3. 歸還掛單扣減的量
		if err := deps.Listing.RestoreAmountInTx(ctx, session, o.ListingID, o.CryptoAmount); err != nil {
			return fmt.Errorf("restore listing amount: %w", err)
		}

		// 4. 寫入訂單狀態日誌
		fromStatus := "matched"
		return deps.StatusLog.AppendInTx(ctx, session, &model.OrderStatusLog{
			OrderID:      o.ID,
			FromStatus:   &fromStatus,
			ToStatus:     "timeout",
			OperatorType: "system",
		})
	})
}
