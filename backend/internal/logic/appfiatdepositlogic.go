package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
	"p2p-exchange/pkg/ecpay"
)

const (
	minFiatDeposit = 100  // 最低 100 TWD
	maxFiatDeposit = 50000 // 最高 50000 TWD
)

type AppFiatDepositLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppFiatDepositLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppFiatDepositLogic {
	return &AppFiatDepositLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppFiatDepositLogic) CreateDeposit(userID int64, req *types.FiatDepositRequest) (*types.FiatDepositResponse, error) {
	conf := l.svcCtx.Config.ECPay
	if !conf.IsEnabled() {
		return nil, apperrors.New(503, "TWD 入金功能尚未開放")
	}
	if req.Amount < minFiatDeposit || req.Amount > maxFiatDeposit {
		return nil, apperrors.New(400, fmt.Sprintf("入金金額須在 %d ~ %d TWD 之間", minFiatDeposit, maxFiatDeposit))
	}

	// Generate unique merchant trade number (≤ 20 chars)
	// Format: P + 15-digit timestamp (nanoseconds / 100)
	tradeNo := fmt.Sprintf("P%015d", time.Now().UnixNano()/100)

	// Create fiat_deposit record
	d := &model.FiatDeposit{
		UserID:          userID,
		Currency:        "TWD",
		Amount:          fmt.Sprintf("%d", req.Amount),
		MerchantTradeNo: tradeNo,
	}
	if err := l.svcCtx.FiatDeposit.Create(l.ctx, d); err != nil {
		return nil, err
	}

	// Build ECPay form parameters
	tradeDate := time.Now().Format("2006/01/02 15:04:05")
	params := ecpay.AioCheckOutParams(
		conf.MerchantID,
		tradeNo,
		tradeDate,
		req.Amount,
		"P2P交易平台TWD入金",
		fmt.Sprintf("TWD入金 %d 元", req.Amount),
		conf.ReturnURL,
		conf.ClientBackURL,
		conf.HashKey,
		conf.HashIV,
	)

	paymentURL := conf.BaseURL + "/Cashier/AioCheckOut/V5"

	return &types.FiatDepositResponse{
		MerchantTradeNo: tradeNo,
		PaymentURL:      paymentURL,
		FormParams:      params,
	}, nil
}

// ── AppListFiatDepositsLogic ──────────────────────────────────────────────────

type AppListFiatDepositsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListFiatDepositsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListFiatDepositsLogic {
	return &AppListFiatDepositsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListFiatDepositsLogic) List(userID int64, req *types.AppListFiatDepositsRequest) (*types.AppListFiatDepositsResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := l.svcCtx.FiatDeposit.ListByUserID(l.ctx, userID, limit, req.Offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	total, err := l.svcCtx.FiatDeposit.CountByUserID(l.ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	items := make([]types.FiatDepositItem, 0, len(rows))
	for _, r := range rows {
		item := types.FiatDepositItem{
			ID:              r.ID,
			Currency:        r.Currency,
			Amount:          r.Amount,
			MerchantTradeNo: r.MerchantTradeNo,
			Status:          r.Status,
			PaymentType:     r.PaymentType,
			CreatedAt:       r.CreatedAt.Format(time.RFC3339),
		}
		if r.PaidAt != nil {
			s := r.PaidAt.Format(time.RFC3339)
			item.PaidAt = &s
		}
		items = append(items, item)
	}

	return &types.AppListFiatDepositsResponse{List: items, Total: total}, nil
}
