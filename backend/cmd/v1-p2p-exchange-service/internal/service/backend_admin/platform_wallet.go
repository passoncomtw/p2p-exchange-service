package backend_admin_service

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	"p2p-exchange/pkg/tron"
)

// GetPlatformWalletInfo 聚合平台錢包資訊：鏈上熱錢包餘額 + DB 內平台虛擬餘額 + ECPay 設定。
// 容錯策略與 legacy 一致：任一來源查詢失敗只記 log 並以 "0" 呈現，不讓後台頁面整頁掛掉。
func (s *backendAdminService) GetPlatformWalletInfo(ctx context.Context) (*backend_interface.PlatformWalletInfoResponse, error) {
	tronConf := s.cfg.Tron
	ecpayConf := s.cfg.ECPay

	onChainBalance := "0"
	if tronConf.IsEnabled() {
		client := tron.NewClient(tronConf.TronGridURL, tronConf.TronGridAPIKey)
		if balance, err := client.GetTRC20Balance(ctx, tronConf.HotWalletAddress, tronConf.USDTContractAddress); err == nil {
			onChainBalance = balance
		} else {
			logx.WithContext(ctx).Errorf("[platform-wallet] GetTRC20Balance: %v", err)
		}
	}

	virtualBalance := "0"
	if balance, err := s.walletRepo.SumCurrencyBalance(ctx, "USDT"); err == nil {
		virtualBalance = balance
	} else {
		logx.WithContext(ctx).Errorf("[platform-wallet] SumCurrencyBalance: %v", err)
	}

	return &backend_interface.PlatformWalletInfoResponse{
		Tron: backend_interface.TronWalletInfo{
			Enabled:                tronConf.IsEnabled(),
			Network:                tronConf.Network,
			HotWalletAddress:       tronConf.HotWalletAddress,
			USDTContractAddress:    tronConf.USDTContractAddress,
			ConfirmationBlocks:     tronConf.ConfirmationBlocks,
			OnChainBalance:         onChainBalance,
			PlatformVirtualBalance: virtualBalance,
		},
		ECPay: backend_interface.ECPayInfo{
			Enabled:    ecpayConf.IsEnabled(),
			MerchantID: ecpayConf.MerchantID,
			BaseURL:    ecpayConf.BaseURL,
		},
	}, nil
}
