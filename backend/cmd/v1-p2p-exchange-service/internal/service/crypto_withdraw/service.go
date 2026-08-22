package cryptowithdraw_service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	cryptowithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_withdraw"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/pkg/tron"
)

const (
	// minUSDTWithdraw 單筆最低提領金額（USDT），與 legacy 一致。
	minUSDTWithdraw = "10"

	// 預設分頁筆數與上限，與 legacy 一致（超出上限一律退回預設值）。
	defaultWithdrawPageSize = 20
	maxWithdrawPageSize     = 100

	usdtCurrency = "USDT"

	statusPending = "pending"
)

// amountRe 只接受十進位正數金額字串（不接受正負號與科學記號），
// 小數位上限對齊 crypto_withdrawals.amount NUMERIC(38,18)：
// 提前擋掉會被 DB 四捨五入的輸入，確保「凍結的金額」與「單據金額」是同一個字串。
var amountRe = regexp.MustCompile(`^\d{1,20}(\.\d{1,18})?$`)

// CryptoWithdrawService 提供 App 端鏈上 USDT 提領申請與查詢。
type CryptoWithdrawService interface {
	// RequestWithdraw 凍結使用者 USDT 餘額並建立 pending 提領申請（同一交易內完成）。
	// 實際的鏈上廣播由 TronWithdrawRunner 非同步處理。
	RequestWithdraw(ctx context.Context, uid int64, toAddress, amount string) (*app_interface.CryptoWithdrawResponse, error)
	// ListWithdrawals 分頁回傳使用者的提領記錄。
	ListWithdrawals(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListCryptoWithdrawalsResponse, error)
}

type cryptoWithdrawService struct {
	db         sqlx.SqlConn
	walletRepo walletrepo.WalletRepository
	repo       cryptowithdrawrepo.CryptoWithdrawRepository
}

func New(
	db sqlx.SqlConn,
	walletRepo walletrepo.WalletRepository,
	repo cryptowithdrawrepo.CryptoWithdrawRepository,
) CryptoWithdrawService {
	return &cryptoWithdrawService{db: db, walletRepo: walletRepo, repo: repo}
}

func (s *cryptoWithdrawService) RequestWithdraw(ctx context.Context, uid int64, toAddress, amount string) (*app_interface.CryptoWithdrawResponse, error) {
	toAddress = strings.TrimSpace(toAddress)
	amount = strings.TrimSpace(amount)

	if toAddress == "" {
		return nil, apierrors.New(400, "toAddress 為必填")
	}
	if amount == "" {
		return nil, apierrors.New(400, "amount 為必填")
	}
	// 位址格式（含 Base58Check 檢查碼）先驗過，避免把資金鎖進一筆註定廣播失敗的申請。
	if _, err := tron.TronBase58ToBytes(toAddress); err != nil {
		return nil, apierrors.New(400, "無效的 Tron 地址")
	}
	if !amountRe.MatchString(amount) {
		return nil, apierrors.New(400, "無效的提領金額")
	}

	parsed, _, err := new(big.Float).Parse(amount, 10)
	if err != nil || parsed.Sign() <= 0 {
		return nil, apierrors.New(400, "無效的提領金額")
	}
	minAmount, _, err := new(big.Float).Parse(minUSDTWithdraw, 10)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	if parsed.Cmp(minAmount) < 0 {
		return nil, apierrors.New(400, fmt.Sprintf("最低提領金額為 %s USDT", minUSDTWithdraw))
	}

	var withdrawID int64
	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先凍結再建單：凍結失敗（餘額不足）時整筆交易回滾，不會留下無資金擔保的提領申請；
		// 建單失敗也會連同凍結一起回滾，不需要 legacy 那種 best-effort 的交易外補償解凍。
		if err := s.walletRepo.FreezeInTx(ctx, session, uid, usdtCurrency, amount); err != nil {
			return err
		}

		created, err := s.repo.CreateInTx(ctx, session, &entity.CryptoWithdrawal{
			UserID:    uid,
			Currency:  usdtCurrency,
			Amount:    amount,
			ToAddress: toAddress,
		})
		if err != nil {
			return apierrors.ErrInternal
		}
		withdrawID = created.ID
		return nil
	}); err != nil {
		// 已具語意的錯誤（例如餘額不足 400）原樣回傳；其餘 DB 錯誤一律收斂成 500，
		// 不把底層錯誤訊息帶到 API 回應。
		var appErr *apierrors.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, apierrors.ErrInternal
	}

	return &app_interface.CryptoWithdrawResponse{ID: withdrawID, Status: statusPending}, nil
}

func (s *cryptoWithdrawService) ListWithdrawals(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListCryptoWithdrawalsResponse, error) {
	if limit <= 0 || limit > maxWithdrawPageSize {
		limit = defaultWithdrawPageSize
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

	items := make([]app_interface.CryptoWithdrawalItem, 0, len(rows))
	for _, r := range rows {
		item := app_interface.CryptoWithdrawalItem{
			ID:        r.ID,
			Currency:  r.Currency,
			Amount:    r.Amount,
			ToAddress: r.ToAddress,
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

	return &app_interface.AppListCryptoWithdrawalsResponse{List: items, Total: total}, nil
}
