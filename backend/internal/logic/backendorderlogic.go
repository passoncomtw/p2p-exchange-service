package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
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
		if err := l.svcCtx.Order.UpdateStatus(l.ctx, req.ID, "completed", map[string]interface{}{
			"completed_at": now,
		}); err != nil {
			return nil, apierrors.ErrInternal
		}

		fromStatus := order.Status
		escrow := &model.EscrowRecord{
			OrderID:        req.ID,
			CryptoCurrency: order.CryptoCurrency,
			Amount:         order.CryptoAmount,
			Action:         "release",
			Status:         "completed",
		}
		if _, err := l.svcCtx.EscrowRecord.Create(l.ctx, escrow); err != nil {
			return nil, apierrors.ErrInternal
		}

		toStatus := "completed"
		if err := l.svcCtx.OrderStatusLog.Append(l.ctx, &model.OrderStatusLog{
			OrderID:      req.ID,
			FromStatus:   &fromStatus,
			ToStatus:     toStatus,
			OperatorType: operatorType,
		}); err != nil {
			return nil, apierrors.ErrInternal
		}

	case "refund":
		reason := req.Reason
		if err := l.svcCtx.Order.UpdateStatus(l.ctx, req.ID, "cancelled", map[string]interface{}{
			"cancelled_at":  now,
			"cancel_reason": reason,
		}); err != nil {
			return nil, apierrors.ErrInternal
		}

		fromStatus := order.Status
		escrow := &model.EscrowRecord{
			OrderID:        req.ID,
			CryptoCurrency: order.CryptoCurrency,
			Amount:         order.CryptoAmount,
			Action:         "refund",
			Status:         "completed",
		}
		if _, err := l.svcCtx.EscrowRecord.Create(l.ctx, escrow); err != nil {
			return nil, apierrors.ErrInternal
		}

		toStatus := "cancelled"
		if err := l.svcCtx.OrderStatusLog.Append(l.ctx, &model.OrderStatusLog{
			OrderID:      req.ID,
			FromStatus:   &fromStatus,
			ToStatus:     toStatus,
			OperatorType: operatorType,
			Remark:       &reason,
		}); err != nil {
			return nil, apierrors.ErrInternal
		}

	default:
		return nil, apierrors.New(400, "invalid action")
	}

	return map[string]interface{}{"ok": true}, nil
}
