package cryptodeposit_service

import (
	"context"
	"fmt"
	"time"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	cryptodepositrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_deposit"
	apierrors "p2p-exchange/internal/errors"
)

const (
	// 預設分頁筆數與上限，與 legacy 一致（超出上限一律退回預設值）。
	defaultDepositPageSize = 20
	maxDepositPageSize     = 100
)

// CryptoDepositService 提供 App 端鏈上 USDT 充值資訊與記錄查詢。
type CryptoDepositService interface {
	// GetDepositInfo 回傳平台熱錢包地址與該使用者專屬的 memo。
	// Tron 未設定（熱錢包地址/私鑰未注入）時回傳 503。
	GetDepositInfo(ctx context.Context, uid int64) (*app_interface.GetCryptoDepositInfoResponse, error)
	// ListDeposits 分頁回傳使用者的充值記錄。
	ListDeposits(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListCryptoDepositsResponse, error)
}

type cryptoDepositService struct {
	cfg  *config.Config
	repo cryptodepositrepo.CryptoDepositRepository
}

func New(cfg *config.Config, repo cryptodepositrepo.CryptoDepositRepository) CryptoDepositService {
	return &cryptoDepositService{cfg: cfg, repo: repo}
}

func (s *cryptoDepositService) GetDepositInfo(ctx context.Context, uid int64) (*app_interface.GetCryptoDepositInfoResponse, error) {
	conf := s.cfg.Tron
	if !conf.IsEnabled() {
		return nil, apierrors.New(503, "USDT 充值功能尚未開放")
	}

	return &app_interface.GetCryptoDepositInfoResponse{
		Address:         conf.HotWalletAddress,
		Memo:            depositMemo(uid),
		Network:         conf.Network,
		Currency:        usdtCurrency,
		ContractAddress: conf.USDTContractAddress,
	}, nil
}

// depositMemo 產生使用者專屬的充值 memo：8 位小寫十六進位的 userID，
// 掃描 Job 以同一組格式反解回 userID（見 scanner_runner.go 的 parseMemoToUserID）。
func depositMemo(uid int64) string {
	return fmt.Sprintf("%08x", uid)
}

func (s *cryptoDepositService) ListDeposits(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListCryptoDepositsResponse, error) {
	if limit <= 0 || limit > maxDepositPageSize {
		limit = defaultDepositPageSize
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.repo.ListByUserID(ctx, uid, limit, offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	total, err := s.repo.CountByUserID(ctx, uid)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]app_interface.CryptoDepositItem, 0, len(rows))
	for _, r := range rows {
		item := app_interface.CryptoDepositItem{
			ID:        r.ID,
			Currency:  r.Currency,
			Amount:    r.Amount,
			TxHash:    r.TxHash,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		}
		if r.ConfirmedAt != nil {
			confirmedAt := r.ConfirmedAt.Format(time.RFC3339)
			item.ConfirmedAt = &confirmedAt
		}
		items = append(items, item)
	}

	return &app_interface.AppListCryptoDepositsResponse{List: items, Total: total}, nil
}
