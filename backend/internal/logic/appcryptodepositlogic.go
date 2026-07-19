package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppGetCryptoDepositInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppGetCryptoDepositInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetCryptoDepositInfoLogic {
	return &AppGetCryptoDepositInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppGetCryptoDepositInfoLogic) GetDepositInfo(userID int64) (*types.GetCryptoDepositInfoResponse, error) {
	conf := l.svcCtx.Config.Tron
	if !conf.IsEnabled() {
		return nil, fmt.Errorf("USDT 充值功能尚未開放")
	}
	// Memo = 8-char zero-padded hex of user ID, used to identify the depositor on the hot wallet
	memo := fmt.Sprintf("%08x", userID)
	return &types.GetCryptoDepositInfoResponse{
		Address:         conf.HotWalletAddress,
		Memo:            memo,
		Network:         conf.Network,
		Currency:        "USDT",
		ContractAddress: conf.USDTContractAddress,
	}, nil
}
