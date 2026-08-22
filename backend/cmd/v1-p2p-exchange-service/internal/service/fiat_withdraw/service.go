package fiatwithdraw_service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	fiatwithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/fiat_withdraw"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/pkg/mq"
	pkgws "p2p-exchange/pkg/ws"
)

const (
	// 單筆提領金額上下限（TWD），與 legacy 一致。
	minFiatWithdraw = 100
	maxFiatWithdraw = 100000

	// 預設分頁筆數與上限，與 legacy 一致（超出上限一律退回預設值）。
	defaultWithdrawPageSize = 20
	maxWithdrawPageSize     = 100

	fiatCurrency = "TWD"
)

var (
	bankCodeRe    = regexp.MustCompile(`^\d{3}$`)
	bankAccountRe = regexp.MustCompile(`^\d{6,20}$`)
)

// FiatWithdrawService 提供 App 端 TWD 提領申請與查詢。
type FiatWithdrawService interface {
	// RequestWithdraw 凍結使用者餘額並建立 pending 提領申請（同一交易內完成），
	// 成功後非同步通知後台。
	RequestWithdraw(ctx context.Context, uid int64, amount, bankCode, bankAccount, accountName string) (*app_interface.FiatWithdrawResponse, error)
	// ListWithdrawals 分頁回傳使用者的提領記錄（銀行帳號已遮蔽）。
	ListWithdrawals(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListFiatWithdrawalsResponse, error)
}

type fiatWithdrawService struct {
	db         sqlx.SqlConn
	walletRepo walletrepo.WalletRepository
	repo       fiatwithdrawrepo.FiatWithdrawRepository
	publisher  mq.Publisher
}

func New(
	db sqlx.SqlConn,
	walletRepo walletrepo.WalletRepository,
	repo fiatwithdrawrepo.FiatWithdrawRepository,
	publisher mq.Publisher,
) FiatWithdrawService {
	return &fiatWithdrawService{db: db, walletRepo: walletRepo, repo: repo, publisher: publisher}
}

func (s *fiatWithdrawService) RequestWithdraw(ctx context.Context, uid int64, amount, bankCode, bankAccount, accountName string) (*app_interface.FiatWithdrawResponse, error) {
	if amount == "" || bankCode == "" || bankAccount == "" || accountName == "" {
		return nil, apierrors.New(400, "amount / bankCode / bankAccount / accountName 為必填")
	}

	parsed, err := strconv.ParseFloat(amount, 64)
	if err != nil || parsed <= 0 {
		return nil, apierrors.New(400, "無效的提領金額")
	}
	if parsed < minFiatWithdraw || parsed > maxFiatWithdraw {
		return nil, apierrors.New(400, fmt.Sprintf("提領金額須在 %d ~ %d TWD 之間", minFiatWithdraw, maxFiatWithdraw))
	}
	if !bankCodeRe.MatchString(bankCode) {
		return nil, apierrors.New(400, "銀行代碼格式錯誤（應為 3 位數字）")
	}
	if !bankAccountRe.MatchString(bankAccount) {
		return nil, apierrors.New(400, "銀行帳號格式錯誤（6~20 位數字）")
	}

	// 統一格式化為兩位小數：與 fiat_withdrawals.amount NUMERIC(18,2) 一致，
	// 且凍結金額與單據金額必定是同一個字串，不會有四捨五入落差。
	amountStr := fmt.Sprintf("%.2f", parsed)

	var withdrawID int64
	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先凍結再建單：凍結失敗（餘額不足）時整筆交易回滾，不會留下無資金擔保的提領申請。
		if err := s.walletRepo.FreezeInTx(ctx, session, uid, fiatCurrency, amountStr); err != nil {
			return err
		}

		created, err := s.repo.CreateInTx(ctx, session, &entity.FiatWithdrawal{
			UserID:      uid,
			Currency:    fiatCurrency,
			Amount:      amountStr,
			BankCode:    bankCode,
			BankAccount: bankAccount,
			AccountName: accountName,
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

	s.notifyAdmin(withdrawID, uid, amountStr)

	return &app_interface.FiatWithdrawResponse{ID: withdrawID, Status: "pending"}, nil
}

// notifyAdmin 在交易成功後（交易外）廣播提領申請通知給後台。
// 通知只是輔助訊息，不是資金異動：發送失敗只記錄 log，不影響已建立的提領申請。
func (s *fiatWithdrawService) notifyAdmin(withdrawID, uid int64, amountStr string) {
	data, err := pkgws.NewMessage(pkgws.EventFiatWithdrawalCreated, pkgws.FiatWithdrawalCreatedData{
		WithdrawalID: withdrawID,
		UserID:       uid,
		Amount:       amountStr,
		Currency:     fiatCurrency,
	})
	if err != nil {
		logx.Errorf("fiat-withdraw: build notify message for withdrawal %d error: %v", withdrawID, err)
		return
	}
	if s.publisher == nil {
		return
	}
	s.publisher.PublishAsync(pkgws.SubjectNotifyAdmin, data)
}

func (s *fiatWithdrawService) ListWithdrawals(ctx context.Context, uid int64, limit, offset int64) (*app_interface.AppListFiatWithdrawalsResponse, error) {
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

	items := make([]app_interface.AppFiatWithdrawalItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, app_interface.AppFiatWithdrawalItem{
			ID:              r.ID,
			Currency:        r.Currency,
			Amount:          r.Amount,
			BankCode:        r.BankCode,
			BankAccountTail: maskBankAccount(r.BankAccount),
			Status:          r.Status,
			CreatedAt:       r.CreatedAt.Format(time.RFC3339),
		})
	}

	return &app_interface.AppListFiatWithdrawalsResponse{List: items, Total: total}, nil
}

// maskBankAccount 只保留銀行帳號後 4 碼，其餘以星號取代（長度 ≤ 4 時原樣回傳）。
func maskBankAccount(account string) string {
	account = strings.TrimSpace(account)
	if len(account) <= 4 {
		return account
	}
	return strings.Repeat("*", len(account)-4) + account[len(account)-4:]
}
