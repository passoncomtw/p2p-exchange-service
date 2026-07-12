package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

// ── AppCreateListingLogic ─────────────────────────────────────────────────────

type AppCreateListingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCreateListingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCreateListingLogic {
	return &AppCreateListingLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppCreateListingLogic) Create(uid int64, req *types.CreateListingRequest) (*types.CreateListingResponse, error) {
	if req.Type == "sell" && req.PaymentMethodID == nil {
		return nil, apierrors.New(400, "sell listing requires a payment method")
	}

	if req.Type == "sell" {
		if err := l.svcCtx.Wallet.Freeze(l.ctx, uid, req.CryptoCurrency, req.TotalAmount); err != nil {
			return nil, err
		}
	}

	listing := &model.Listing{
		UserID:           uid,
		Type:             req.Type,
		CryptoCurrency:   req.CryptoCurrency,
		FiatCurrency:     req.FiatCurrency,
		TotalAmount:      req.TotalAmount,
		RemainingAmount:  req.TotalAmount,
		Price:            req.Price,
		MinOrderFiat:     req.MinOrderFiat,
		MaxOrderFiat:     req.MaxOrderFiat,
		PlatformFeeBase:  0,
		PlatformFeeRate:  0,
		PaymentFeeBase:   0,
		PaymentFeeRate:   0,
		PaymentTimeLimit: req.PaymentTimeLimit,
		PaymentMethodID:  req.PaymentMethodID,
		Status:           "active",
	}

	id, err := l.svcCtx.Listing.Create(l.ctx, listing)
	if err != nil {
		l.Errorf("create listing failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	return &types.CreateListingResponse{ID: id}, nil
}

// ── AppListListingsLogic ──────────────────────────────────────────────────────

type AppListListingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListListingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListListingsLogic {
	return &AppListListingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppListListingsLogic) List(req *types.ListListingsRequest) (*types.ListListingsResponse, error) {
	rows, err := l.svcCtx.Listing.List(l.ctx, req.Type, req.Status, req.Limit, req.Offset)
	if err != nil {
		l.Errorf("list listings failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	list := make([]types.ListingItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, listingToItem(row))
	}

	return &types.ListListingsResponse{List: list}, nil
}

// ── AppMyListingsLogic ────────────────────────────────────────────────────────

type AppMyListingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppMyListingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppMyListingsLogic {
	return &AppMyListingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppMyListingsLogic) List(uid int64, req *types.ListListingsRequest) (*types.ListListingsResponse, error) {
	rows, err := l.svcCtx.Listing.ListByUser(l.ctx, uid, req.Type, req.Status, req.Limit, req.Offset)
	if err != nil {
		l.Errorf("list my listings uid=%d failed: %v", uid, err)
		return nil, apierrors.ErrInternal
	}

	list := make([]types.ListingItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, listingToItem(row))
	}

	return &types.ListListingsResponse{List: list}, nil
}

// ── AppGetListingLogic ────────────────────────────────────────────────────────

type AppGetListingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppGetListingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetListingLogic {
	return &AppGetListingLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppGetListingLogic) Get(id int64) (*types.ListingItem, error) {
	row, err := l.svcCtx.Listing.FindByID(l.ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		l.Errorf("get listing id=%d failed: %v", id, err)
		return nil, apierrors.ErrInternal
	}

	item := listingToItem(row)
	return &item, nil
}

// ── AppCancelListingLogic ─────────────────────────────────────────────────────

type AppCancelListingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCancelListingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCancelListingLogic {
	return &AppCancelListingLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppCancelListingLogic) Cancel(uid int64, id int64) error {
	listing, err := l.svcCtx.Listing.FindByID(l.ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		l.Errorf("get listing id=%d failed: %v", id, err)
		return apierrors.ErrInternal
	}

	if listing.UserID != uid {
		return apierrors.ErrForbidden
	}

	if listing.Status != "active" && listing.Status != "paused" {
		return apierrors.New(400, "only active or paused listings can be cancelled")
	}

	if err := l.svcCtx.Listing.UpdateStatus(l.ctx, id, "cancelled"); err != nil {
		l.Errorf("cancel listing id=%d failed: %v", id, err)
		return apierrors.ErrInternal
	}

	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func listingToItem(l *model.Listing) types.ListingItem {
	return types.ListingItem{
		ID:               l.ID,
		UserID:           l.UserID,
		Type:             l.Type,
		CryptoCurrency:   l.CryptoCurrency,
		FiatCurrency:     l.FiatCurrency,
		TotalAmount:      l.TotalAmount,
		RemainingAmount:  l.RemainingAmount,
		Price:            l.Price,
		MinOrderFiat:     l.MinOrderFiat,
		MaxOrderFiat:     l.MaxOrderFiat,
		PlatformFeeBase:  l.PlatformFeeBase,
		PlatformFeeRate:  l.PlatformFeeRate,
		PaymentFeeBase:   l.PaymentFeeBase,
		PaymentFeeRate:   l.PaymentFeeRate,
		PaymentTimeLimit: l.PaymentTimeLimit,
		PaymentMethodID:  l.PaymentMethodID,
		Status:           l.Status,
		CreatedAt:        l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
