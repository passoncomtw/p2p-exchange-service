package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
	"p2p-exchange/pkg/tron"
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
	tronConf := l.svcCtx.Config.Tron
	ecpay := l.svcCtx.Config.ECPay

	onChainBalance := "0"
	if tronConf.IsEnabled() {
		client := tron.NewClient(tronConf.TronGridURL, tronConf.TronGridAPIKey)
		if bal, err := client.GetTRC20Balance(l.ctx, tronConf.HotWalletAddress, tronConf.USDTContractAddress); err == nil {
			onChainBalance = bal
		} else {
			l.Errorf("[platform-wallet] GetTRC20Balance: %v", err)
		}
	}

	virtualBalance := "0"
	if bal, err := l.svcCtx.Wallet.SumCurrencyBalance(l.ctx, "USDT"); err == nil {
		virtualBalance = bal
	} else {
		l.Errorf("[platform-wallet] SumCurrencyBalance: %v", err)
	}

	return &types.PlatformWalletInfoResponse{
		Tron: types.TronWalletInfo{
			Enabled:                tronConf.IsEnabled(),
			Network:                tronConf.Network,
			HotWalletAddress:       tronConf.HotWalletAddress,
			USDTContractAddress:    tronConf.USDTContractAddress,
			ConfirmationBlocks:     tronConf.ConfirmationBlocks,
			OnChainBalance:         onChainBalance,
			PlatformVirtualBalance: virtualBalance,
		},
		ECPay: types.ECPayInfo{
			Enabled:    ecpay.IsEnabled(),
			MerchantID: ecpay.MerchantID,
			BaseURL:    ecpay.BaseURL,
		},
	}, nil
}
