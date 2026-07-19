package logic

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppFiatWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppFiatWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppFiatWithdrawLogic {
	return &AppFiatWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppFiatWithdrawLogic) RequestWithdraw(userID int64, req *types.FiatWithdrawRequest) (*types.FiatWithdrawResponse, error) {
	if req.Amount == "" || req.BankCode == "" || req.BankAccount == "" || req.AccountName == "" {
		return nil, apperrors.New(400, "amount / bankCode / bankAccount / accountName 為必填")
	}

	// Validate amount is positive
	amount, _, err := new(big.Float).Parse(req.Amount, 10)
	if err != nil || amount.Sign() <= 0 {
		return nil, apperrors.New(400, "無效的提領金額")
	}

	// Freeze TWD balance
	amountF, _ := amount.Float64()
	if err := l.svcCtx.Wallet.Freeze(l.ctx, userID, "TWD", amountF); err != nil {
		return nil, err
	}

	// Create fiat_withdrawal record (pending human review)
	w := &model.FiatWithdrawal{
		UserID:      userID,
		Currency:    "TWD",
		Amount:      req.Amount,
		BankCode:    req.BankCode,
		BankAccount: req.BankAccount,
		AccountName: req.AccountName,
	}
	if err := l.svcCtx.FiatWithdraw.Create(l.ctx, w); err != nil {
		_ = l.svcCtx.Wallet.UnfreezeBalance(l.ctx, userID, "TWD", req.Amount)
		return nil, err
	}

	return &types.FiatWithdrawResponse{
		ID:     w.ID,
		Status: "pending",
	}, nil
}
