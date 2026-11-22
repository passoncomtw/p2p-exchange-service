package fiatdeposit_service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	fiatdepositrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/fiat_deposit"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/money"
	"p2p-exchange/pkg/ecpay"
)

// 單筆入金金額上下限（TWD），與 legacy 一致。
var (
	minFiatDeposit = decimal.NewFromInt(100)
	maxFiatDeposit = decimal.NewFromInt(50000)
)

const (
	// 預設分頁筆數與上限，與 legacy 一致（超出上限一律退回預設值）。
	defaultDepositPageSize = 20
	maxDepositPageSize     = 100
)

// FiatDepositService 提供 App 端 TWD 入金（ECPay）建立與查詢。
type FiatDepositService interface {
	// CreateDeposit 建立一筆 pending 入金記錄並回傳 ECPay 導轉表單參數。
	// ECPay 未設定（HashKey/HashIV 未注入）時回傳 503。
	CreateDeposit(ctx context.Context, uid int64, amount string) (*app_interface.FiatDepositResponse, error)
	// ListDeposits 分頁回傳使用者的入金記錄。
	ListDeposits(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListFiatDepositsResponse, error)
}

type fiatDepositService struct {
	cfg  *config.Config
	repo fiatdepositrepo.FiatDepositRepository
}

func New(cfg *config.Config, repo fiatdepositrepo.FiatDepositRepository) FiatDepositService {
	return &fiatDepositService{cfg: cfg, repo: repo}
}

func (s *fiatDepositService) CreateDeposit(ctx context.Context, uid int64, amountInput string) (*app_interface.FiatDepositResponse, error) {
	conf := s.cfg.ECPay
	if !conf.IsEnabled() {
		return nil, apierrors.New(503, "TWD 入金功能尚未開放")
	}
	amount, err := money.Parse(amountInput)
	if err != nil || !amount.IsPositive() {
		return nil, apierrors.New(400, "無效的入金金額")
	}
	if amount.LessThan(minFiatDeposit) || amount.GreaterThan(maxFiatDeposit) {
		return nil, apierrors.New(400, fmt.Sprintf("入金金額須在 %s ~ %s TWD 之間", minFiatDeposit, maxFiatDeposit))
	}
	// ECPay AioCheckOut 的 TotalAmount 只接受整數元，小數金額無法交由金流處理。
	if !amount.IsInteger() {
		return nil, apierrors.New(400, "入金金額須為整數元")
	}

	// MerchantTradeNo 需 ≤ 20 字元且全平台唯一：P + 15 位時間戳（奈秒 / 100）。
	tradeNo := fmt.Sprintf("P%015d", time.Now().UnixNano()/100)

	// 先落地 pending 記錄再產生導轉參數：若寫入失敗就不會有無主的 ECPay 訂單。
	d := &entity.FiatDeposit{
		UserID:          uid,
		Currency:        "TWD",
		Amount:          amount.String(),
		MerchantTradeNo: tradeNo,
		Status:          "pending",
	}
	if _, err := s.repo.Create(ctx, d); err != nil {
		return nil, apierrors.ErrInternal
	}

	tradeDate := time.Now().Format("2006/01/02 15:04:05")
	params := ecpay.AioCheckOutParams(
		conf.MerchantID,
		tradeNo,
		tradeDate,
		int(amount.IntPart()),
		"P2P交易平台TWD入金",
		fmt.Sprintf("TWD入金 %s 元", amount),
		conf.ReturnURL,
		conf.ClientBackURL,
		conf.HashKey,
		conf.HashIV,
	)

	return &app_interface.FiatDepositResponse{
		MerchantTradeNo: tradeNo,
		PaymentURL:      conf.BaseURL + "/Cashier/AioCheckOut/V5",
		FormParams:      params,
	}, nil
}

func (s *fiatDepositService) ListDeposits(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListFiatDepositsResponse, error) {
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

	items := make([]app_interface.FiatDepositItem, 0, len(rows))
	for _, r := range rows {
		item := app_interface.FiatDepositItem{
			ID:              r.ID,
			Currency:        r.Currency,
			Amount:          r.Amount,
			MerchantTradeNo: r.MerchantTradeNo,
			Status:          r.Status,
			PaymentType:     r.PaymentType,
			CreatedAt:       r.CreatedAt.Format(time.RFC3339),
		}
		if r.PaidAt != nil {
			paidAt := r.PaidAt.Format(time.RFC3339)
			item.PaidAt = &paidAt
		}
		items = append(items, item)
	}

	return &app_interface.AppListFiatDepositsResponse{List: items, Total: total}, nil
}
