package wallet_service

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

// WalletService 提供 App 端錢包唯讀查詢（餘額與帳本流水）。
type WalletService interface {
	// ListWallets 回傳使用者所有幣別的錢包餘額。
	ListWallets(ctx context.Context, uid int64) (*app_interface.ListWalletsResponse, error)
	// ListLedgers 分頁回傳使用者指定幣別的帳本流水；該幣別錢包不存在時回傳 404。
	ListLedgers(ctx context.Context, uid int64, currency string, limit, offset int64) (*app_interface.ListWalletLedgersResponse, error)
}

type walletService struct {
	walletRepo walletrepo.WalletRepository
}

func New(walletRepo walletrepo.WalletRepository) WalletService {
	return &walletService{walletRepo: walletRepo}
}

func (s *walletService) ListWallets(ctx context.Context, uid int64) (*app_interface.ListWalletsResponse, error) {
	wallets, err := s.walletRepo.FindByUserID(ctx, uid)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	list := make([]app_interface.WalletItem, 0, len(wallets))
	for _, w := range wallets {
		list = append(list, app_interface.WalletItem{
			Currency:         w.Currency,
			AvailableBalance: w.AvailableBalance,
			FrozenBalance:    w.FrozenBalance,
		})
	}
	return &app_interface.ListWalletsResponse{List: list}, nil
}

func (s *walletService) ListLedgers(ctx context.Context, uid int64, currency string, limit, offset int64) (*app_interface.ListWalletLedgersResponse, error) {
	// 先以 uid + currency 定位錢包，後續帳本查詢一律以查出的 wallet.ID 為準，
	// 避免用呼叫端傳入的 walletID 直接查詢而讀到他人帳本。
	wallet, err := s.walletRepo.FindOne(ctx, uid, currency)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		return nil, apierrors.ErrInternal
	}

	ledgers, total, err := s.walletRepo.ListLedgers(ctx, wallet.ID, int(limit), int(offset))
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	list := make([]app_interface.WalletLedgerItem, 0, len(ledgers))
	for _, l := range ledgers {
		list = append(list, app_interface.WalletLedgerItem{
			Type:         l.Type,
			Amount:       l.Amount,
			BalanceAfter: l.BalanceAfter,
			RefOrderNo:   l.RefOrderNo,
			CreatedAt:    l.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return &app_interface.ListWalletLedgersResponse{List: list, Total: total}, nil
}
