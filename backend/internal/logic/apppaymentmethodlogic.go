package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

// ── AppCreatePaymentMethodLogic ───────────────────────────────────────────────

type AppCreatePaymentMethodLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCreatePaymentMethodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCreatePaymentMethodLogic {
	return &AppCreatePaymentMethodLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppCreatePaymentMethodLogic) Create(uid int64, req *types.CreatePaymentMethodRequest) (*types.CreatePaymentMethodResponse, error) {
	pm := &model.PaymentMethod{
		UserID:        uid,
		Type:          req.Type,
		BankName:      req.BankName,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		IsActive:      true,
	}

	id, err := l.svcCtx.PaymentMethod.Create(l.ctx, pm)
	if err != nil {
		l.Errorf("create payment method failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	return &types.CreatePaymentMethodResponse{ID: id}, nil
}

// ── AppListPaymentMethodsLogic ────────────────────────────────────────────────

type AppListPaymentMethodsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListPaymentMethodsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListPaymentMethodsLogic {
	return &AppListPaymentMethodsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppListPaymentMethodsLogic) List(uid int64) (*types.ListPaymentMethodsResponse, error) {
	rows, err := l.svcCtx.PaymentMethod.FindByUserID(l.ctx, uid)
	if err != nil {
		l.Errorf("list payment methods failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	list := make([]types.PaymentMethodItem, 0, len(rows))
	for _, pm := range rows {
		list = append(list, types.PaymentMethodItem{
			ID:            pm.ID,
			Type:          pm.Type,
			BankName:      pm.BankName,
			AccountName:   pm.AccountName,
			AccountNumber: pm.AccountNumber,
			IsActive:      pm.IsActive,
		})
	}

	return &types.ListPaymentMethodsResponse{List: list}, nil
}

// ── AppDeletePaymentMethodLogic ───────────────────────────────────────────────

type AppDeletePaymentMethodLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppDeletePaymentMethodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppDeletePaymentMethodLogic {
	return &AppDeletePaymentMethodLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppDeletePaymentMethodLogic) Delete(uid int64, id int64) error {
	err := l.svcCtx.PaymentMethod.SoftDelete(l.ctx, id, uid)
	if err != nil {
		l.Errorf("delete payment method id=%d uid=%d failed: %v", id, uid, err)
		return apierrors.ErrInternal
	}
	return nil
}
