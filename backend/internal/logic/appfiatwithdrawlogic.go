package logic

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

const (
	minFiatWithdraw = 100
	maxFiatWithdraw = 100000
)

var (
	bankCodeRe    = regexp.MustCompile(`^\d{3}$`)
	bankAccountRe = regexp.MustCompile(`^\d{6,20}$`)
)

type AppFiatWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppFiatWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppFiatWithdrawLogic {
	return &AppFiatWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppFiatWithdrawLogic) RequestWithdraw(userID int64, req *types.FiatWithdrawRequest) (*types.FiatWithdrawResponse, error) {
	if req.Amount == "" || req.BankCode == "" || req.BankAccount == "" || req.AccountName == "" {
		return nil, apperrors.New(400, "amount / bankCode / bankAccount / accountName 為必填")
	}

	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		return nil, apperrors.New(400, "無效的提領金額")
	}
	if amount < minFiatWithdraw || amount > maxFiatWithdraw {
		return nil, apperrors.New(400, fmt.Sprintf("提領金額須在 %d ~ %d TWD 之間", minFiatWithdraw, maxFiatWithdraw))
	}

	if !bankCodeRe.MatchString(req.BankCode) {
		return nil, apperrors.New(400, "銀行代碼格式錯誤（應為 3 位數字）")
	}
	if !bankAccountRe.MatchString(req.BankAccount) {
		return nil, apperrors.New(400, "銀行帳號格式錯誤（6~20 位數字）")
	}

	amountStr := fmt.Sprintf("%.2f", amount)

	var withdrawID int64
	if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.Wallet.FreezeInTx(ctx, session, userID, "TWD", amountStr); err != nil {
			return err
		}
		w := &model.FiatWithdrawal{
			UserID:      userID,
			Currency:    "TWD",
			Amount:      amountStr,
			BankCode:    req.BankCode,
			BankAccount: req.BankAccount,
			AccountName: req.AccountName,
		}
		if err := l.svcCtx.FiatWithdraw.CreateInTx(ctx, session, w); err != nil {
			return err
		}
		withdrawID = w.ID
		return nil
	}); err != nil {
		return nil, err
	}

	return &types.FiatWithdrawResponse{
		ID:     withdrawID,
		Status: "pending",
	}, nil
}

// ── AppListFiatWithdrawalsLogic ───────────────────────────────────────────────

type AppListFiatWithdrawalsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListFiatWithdrawalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListFiatWithdrawalsLogic {
	return &AppListFiatWithdrawalsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListFiatWithdrawalsLogic) List(userID int64, req *types.AppListFiatWithdrawalsRequest) (*types.AppListFiatWithdrawalsResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := l.svcCtx.FiatWithdraw.ListByUserID(l.ctx, userID, limit, req.Offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	total, err := l.svcCtx.FiatWithdraw.CountByUserID(l.ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	items := make([]types.AppFiatWithdrawalItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, types.AppFiatWithdrawalItem{
			ID:             r.ID,
			Currency:       r.Currency,
			Amount:         r.Amount,
			BankCode:       r.BankCode,
			BankAccountTail: maskBankAccount(r.BankAccount),
			Status:         r.Status,
			CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.AppListFiatWithdrawalsResponse{List: items, Total: total}, nil
}

// maskBankAccount returns the last 4 digits with leading asterisks.
func maskBankAccount(account string) string {
	account = strings.TrimSpace(account)
	if len(account) <= 4 {
		return account
	}
	return strings.Repeat("*", len(account)-4) + account[len(account)-4:]
}
