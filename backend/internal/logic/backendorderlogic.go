package logic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
	pkgws "p2p-exchange/pkg/ws"
)

// ── BackendListListingsLogic ──────────────────────────────────────────────────

type BackendListListingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendListListingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendListListingsLogic {
	return &BackendListListingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BackendListListingsLogic) List(req *types.BackendListListingsRequest) (*types.ListListingsResponse, error) {
	rows, err := l.svcCtx.Listing.List(l.ctx, req.Type, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]types.ListingItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, types.ListingItem{
			ID:               r.ID,
			UserID:           r.UserID,
			Type:             r.Type,
			CryptoCurrency:   r.CryptoCurrency,
			FiatCurrency:     r.FiatCurrency,
			TotalAmount:      r.TotalAmount,
			RemainingAmount:  r.RemainingAmount,
			Price:            r.Price,
			MinOrderFiat:     r.MinOrderFiat,
			MaxOrderFiat:     r.MaxOrderFiat,
			PlatformFeeBase:  r.PlatformFeeBase,
			PlatformFeeRate:  r.PlatformFeeRate,
			PaymentFeeBase:   r.PaymentFeeBase,
			PaymentFeeRate:   r.PaymentFeeRate,
			PaymentTimeLimit: r.PaymentTimeLimit,
			PaymentMethodID:  r.PaymentMethodID,
			Status:           r.Status,
			CreatedAt:        r.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.ListListingsResponse{List: items}, nil
}

// ── BackendListOrdersLogic ────────────────────────────────────────────────────

type BackendListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendListOrdersLogic {
	return &BackendListOrdersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BackendListOrdersLogic) List(req *types.BackendListOrdersRequest) (*types.BackendListOrdersResponse, error) {
	rows, err := l.svcCtx.Order.BackendList(l.ctx, req.Keyword, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]types.OrderItem, 0, len(rows))
	for _, r := range rows {
		item := types.OrderItem{
			ID:                r.ID,
			OrderNo:           r.OrderNo,
			ListingID:         r.ListingID,
			ListingType:       r.ListingType,
			SellerID:          r.SellerID,
			BuyerID:           r.BuyerID,
			CryptoCurrency:    r.CryptoCurrency,
			FiatCurrency:      r.FiatCurrency,
			CryptoAmount:      r.CryptoAmount,
			Price:             r.Price,
			FiatAmount:        r.FiatAmount,
			PlatformFeeBase:   r.PlatformFeeBase,
			PlatformFeeAmount: r.PlatformFeeAmount,
			PaymentFeeBase:    r.PaymentFeeBase,
			PaymentFeeAmount:  r.PaymentFeeAmount,
			TotalFee:          r.TotalFee,
			TotalAmount:       r.TotalAmount,
			PaymentMethodID:   r.PaymentMethodID,
			Status:            r.Status,
			PaymentDeadline:   r.PaymentDeadline.Format(time.RFC3339),
			CancelReason:      r.CancelReason,
			CreatedAt:         r.CreatedAt.Format(time.RFC3339),
		}
		if r.PaidAt != nil {
			s := r.PaidAt.Format(time.RFC3339)
			item.PaidAt = &s
		}
		if r.ConfirmedAt != nil {
			s := r.ConfirmedAt.Format(time.RFC3339)
			item.ConfirmedAt = &s
		}
		if r.CompletedAt != nil {
			s := r.CompletedAt.Format(time.RFC3339)
			item.CompletedAt = &s
		}
		if r.CancelledAt != nil {
			s := r.CancelledAt.Format(time.RFC3339)
			item.CancelledAt = &s
		}
		items = append(items, item)
	}

	total, err := l.svcCtx.Order.BackendCount(l.ctx, req.Keyword, req.Status)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	return &types.BackendListOrdersResponse{List: items, Total: total}, nil
}

// ── BackendResolveOrderLogic ──────────────────────────────────────────────────

type BackendResolveOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendResolveOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendResolveOrderLogic {
	return &BackendResolveOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BackendResolveOrderLogic) Resolve(req *types.ResolveOrderRequest) (interface{}, error) {
	order, err := l.svcCtx.Order.FindByID(l.ctx, req.ID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	if order.Status != "disputed" {
		return nil, apierrors.New(400, "order is not in disputed status")
	}

	now := time.Now()
	operatorType := "admin"

	switch req.Action {
	case "complete":
		if l.svcCtx.RDB != nil {
			unlockSeller, err := l.svcCtx.RDB.AcquireLock(l.ctx, model.WalletLockKey(order.SellerID, order.CryptoCurrency), 10*time.Second)
			if err != nil {
				return nil, apierrors.ErrInternal
			}
			defer unlockSeller()
			unlockBuyer, err := l.svcCtx.RDB.AcquireLock(l.ctx, model.WalletLockKey(order.BuyerID, order.CryptoCurrency), 10*time.Second)
			if err != nil {
				return nil, apierrors.ErrInternal
			}
			defer unlockBuyer()
		}

		if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			if _, err := session.ExecCtx(ctx,
				`UPDATE orders SET status='completed', completed_at=$1, updated_at=NOW() WHERE id=$2`,
				now, req.ID,
			); err != nil {
				return err
			}

			if _, err := l.svcCtx.EscrowRecord.Create(l.ctx, &model.EscrowRecord{
				OrderID:        req.ID,
				CryptoCurrency: order.CryptoCurrency,
				Amount:         order.CryptoAmount,
				Action:         "release",
				Status:         "completed",
			}); err != nil {
				return err
			}

			if err := l.svcCtx.Wallet.TransferInTx(ctx, session, order.SellerID, order.BuyerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo); err != nil {
				return err
			}

			fromStatus := order.Status
			toStatus := "completed"
			return l.svcCtx.OrderStatusLog.AppendInTx(ctx, session, &model.OrderStatusLog{
				OrderID:      req.ID,
				FromStatus:   &fromStatus,
				ToStatus:     toStatus,
				OperatorType: operatorType,
			})
		}); err != nil {
			return nil, err
		}

	case "refund":
		if l.svcCtx.RDB != nil {
			unlock, err := l.svcCtx.RDB.AcquireLock(l.ctx, model.WalletLockKey(order.SellerID, order.CryptoCurrency), 10*time.Second)
			if err != nil {
				return nil, apierrors.ErrInternal
			}
			defer unlock()
		}

		reason := req.Reason
		if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			if _, err := session.ExecCtx(ctx,
				`UPDATE orders SET status='cancelled', cancelled_at=$1, cancel_reason=$2, updated_at=NOW() WHERE id=$3`,
				now, reason, req.ID,
			); err != nil {
				return err
			}

			if _, err := l.svcCtx.EscrowRecord.Create(l.ctx, &model.EscrowRecord{
				OrderID:        req.ID,
				CryptoCurrency: order.CryptoCurrency,
				Amount:         order.CryptoAmount,
				Action:         "refund",
				Status:         "completed",
			}); err != nil {
				return err
			}

			if err := l.svcCtx.Wallet.UnfreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo); err != nil {
				return err
			}

			if err := l.svcCtx.Listing.RestoreAmountInTx(ctx, session, order.ListingID, order.CryptoAmount); err != nil {
				return err
			}

			fromStatus := order.Status
			toStatus := "cancelled"
			return l.svcCtx.OrderStatusLog.AppendInTx(ctx, session, &model.OrderStatusLog{
				OrderID:      req.ID,
				FromStatus:   &fromStatus,
				ToStatus:     toStatus,
				OperatorType: operatorType,
				Remark:       &reason,
			})
		}); err != nil {
			return nil, err
		}

	default:
		return nil, apierrors.New(400, "invalid action")
	}

	// 推送 WS 事件
	if l.svcCtx.MQ != nil {
		finalStatus := "completed"
		if req.Action == "refund" {
			finalStatus = "cancelled"
		}
		payload := pkgws.OrderStatusChangedData{
			OrderID:  order.ID,
			Status:   finalStatus,
			BuyerID:  order.BuyerID,
			SellerID: order.SellerID,
		}
		if data, err := json.Marshal(payload); err == nil {
			l.svcCtx.MQ.PublishAsync(pkgws.EventOrderStatusChanged, data)
		}
	}

	return map[string]interface{}{"ok": true}, nil
}
