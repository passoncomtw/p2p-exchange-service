package order_service

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go.uber.org/fx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	orderstatuslogrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order_status_log"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/orderallocation"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/pkg/mq"
)

const expiredOrderLockTTL = 30 * time.Second

type ExpiredOrderConsumerDeps struct {
	fx.In

	DB            sqlx.SqlConn
	RDB           *rdb.Client
	Subscriber    mq.Subscriber
	OrderRepo     orderrepo.OrderRepository
	ListingRepo   listingrepo.ListingRepository
	WalletRepo    walletrepo.WalletRepository
	StatusLogRepo orderstatuslogrepo.OrderStatusLogRepository
}

// RegisterExpiredOrderConsumer 訂閱 order.timeout.check，批次取消超時訂單。
func RegisterExpiredOrderConsumer(deps ExpiredOrderConsumerDeps) {
	if err := deps.Subscriber.Subscribe("order.timeout.check", func(ctx context.Context, _ []byte) error {
		return processExpiredOrders(ctx, deps)
	}); err != nil {
		logx.Errorf("expired-order: subscribe error: %v", err)
	}
}

func processExpiredOrders(ctx context.Context, deps ExpiredOrderConsumerDeps) error {
	orders, err := deps.OrderRepo.FindExpired(ctx)
	if err != nil {
		return fmt.Errorf("find expired orders: %w", err)
	}
	for _, o := range orders {
		if err := cancelExpiredOrder(ctx, deps, o); err != nil {
			logx.Errorf("expired-order: cancel order %d error: %v", o.ID, err)
		}
	}
	return nil
}

// cancelExpiredOrder 把逾時的 matched 訂單轉為 timeout，資產處理與使用者取消相同（orderallocation.ReleaseInTx）。
func cancelExpiredOrder(ctx context.Context, deps ExpiredOrderConsumerDeps, order *entity.Order) error {
	if deps.RDB != nil {
		lockKey := fmt.Sprintf("processing:order:%d", order.ID)
		unlock, err := deps.RDB.AcquireLock(ctx, lockKey, expiredOrderLockTTL)
		if err != nil {
			return nil
		}
		defer unlock()
	}

	return deps.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			`UPDATE orders SET status = 'timeout', cancelled_at = NOW(), updated_at = NOW()
			 WHERE id = $1 AND status = 'matched'`,
			order.ID,
		)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}

		if err := orderallocation.ReleaseInTx(ctx, session, deps.ListingRepo, deps.WalletRepo, order); err != nil {
			return fmt.Errorf("release order allocation: %w", err)
		}

		fromStatus := "matched"
		return deps.StatusLogRepo.AppendInTx(ctx, session, &orderstatuslogrepo.OrderStatusLog{
			OrderID:      order.ID,
			FromStatus:   &fromStatus,
			ToStatus:     "timeout",
			OperatorType: "system",
		})
	})
}
