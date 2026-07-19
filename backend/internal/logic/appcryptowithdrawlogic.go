package logic

import (
	"context"
	"fmt"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
	"p2p-exchange/pkg/tron"
)

const minUSDTWithdraw = "10" // minimum 10 USDT

type AppCryptoWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCryptoWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCryptoWithdrawLogic {
	return &AppCryptoWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppCryptoWithdrawLogic) Withdraw(userID int64, req *types.CryptoWithdrawRequest) (*types.CryptoWithdrawResponse, error) {
	if req.ToAddress == "" {
		return nil, apperrors.New(400, "toAddress 為必填")
	}
	if req.Amount == "" {
		return nil, apperrors.New(400, "amount 為必填")
	}

	// Validate Tron address
	if _, err := tron.TronBase58ToBytes(req.ToAddress); err != nil {
		return nil, apperrors.New(400, "無效的 Tron 地址")
	}

	// Validate amount
	amount, _, err := new(big.Float).Parse(req.Amount, 10)
	if err != nil || amount.Sign() <= 0 {
		return nil, apperrors.New(400, "無效的提領金額")
	}
	minAmount, _, _ := new(big.Float).Parse(minUSDTWithdraw, 10)
	if amount.Cmp(minAmount) < 0 {
		return nil, fmt.Errorf("最低提領金額為 %s USDT", minUSDTWithdraw)
	}

	// Freeze balance (deduct from available → frozen)
	amountF, _ := amount.Float64()
	if err := l.svcCtx.Wallet.Freeze(l.ctx, userID, "USDT", amountF); err != nil {
		return nil, err
	}

	// Record withdrawal request
	w := &model.CryptoWithdrawal{
		UserID:    userID,
		Currency:  "USDT",
		Amount:    req.Amount,
		ToAddress: req.ToAddress,
	}
	if err := l.svcCtx.CryptoWithdraw.Create(l.ctx, w); err != nil {
		// Best-effort unfreeze on failure to persist
		_ = l.svcCtx.Wallet.UnfreezeBalance(l.ctx, userID, "USDT", req.Amount)
		return nil, err
	}

	return &types.CryptoWithdrawResponse{
		ID:     w.ID,
		Status: "pending",
	}, nil
}
