package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendPlatformWalletLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendPlatformWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendPlatformWalletLogic {
	return &BackendPlatformWalletLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BackendPlatformWalletLogic) GetWalletInfo() (*types.PlatformWalletInfoResponse, error) {
	tron := l.svcCtx.Config.Tron
	ecpay := l.svcCtx.Config.ECPay

	return &types.PlatformWalletInfoResponse{
		Tron: types.TronWalletInfo{
			Enabled:             tron.IsEnabled(),
			Network:             tron.Network,
			HotWalletAddress:    tron.HotWalletAddress,
			USDTContractAddress: tron.USDTContractAddress,
			ConfirmationBlocks:  tron.ConfirmationBlocks,
		},
		ECPay: types.ECPayInfo{
			Enabled:    ecpay.IsEnabled(),
			MerchantID: ecpay.MerchantID,
			BaseURL:    ecpay.BaseURL,
		},
	}, nil
}
