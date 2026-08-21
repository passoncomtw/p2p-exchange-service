package backend_admin_service

import (
	"context"
	"math/big"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	apierrors "p2p-exchange/internal/errors"
)

func (s *backendAdminService) ListMembers(ctx context.Context, keyword string, limit, offset int64) (*backend_interface.BackendListMembersResponse, error) {
	rows, err := s.userRepo.List(ctx, keyword, limit, offset)
	if err != nil {
		logx.WithContext(ctx).Errorf("[backend-members] list failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	total, err := s.userRepo.Count(ctx, keyword)
	if err != nil {
		logx.WithContext(ctx).Errorf("[backend-members] count failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	items := make([]backend_interface.MemberItem, 0, len(rows))
	for _, u := range rows {
		email := ""
		if u.Email != nil {
			email = *u.Email
		}
		items = append(items, backend_interface.MemberItem{
			ID:        u.ID,
			Username:  u.Username,
			Email:     email,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &backend_interface.BackendListMembersResponse{List: items, Total: total}, nil
}

// DepositToMember 後台手動入金。
// adminUID 為操作人（後台帳號 ID），由 handler 從 JWT context 取得；本方法只負責記錄，
// 不接受任何來自 request body 的操作人資訊。
func (s *backendAdminService) DepositToMember(ctx context.Context, adminUID int64, id int64, currency string, amount string) (*backend_interface.BackendDepositResponse, error) {
	// 金額必須是正數：負數會讓入金變成扣款。
	amt, ok := new(big.Float).SetString(amount)
	if !ok || amt.Sign() <= 0 {
		return nil, apierrors.New(400, "amount must be a positive number")
	}

	if currency == "" {
		return nil, apierrors.New(400, "currency is required")
	}
	exists, err := s.currencyExists(ctx, currency)
	if err != nil {
		logx.WithContext(ctx).Errorf("[backend-deposit] check currency %s failed: %v", currency, err)
		return nil, apierrors.ErrInternal
	}
	if !exists {
		return nil, apierrors.New(400, "unsupported currency")
	}

	if _, err := s.userRepo.FindByID(ctx, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		logx.WithContext(ctx).Errorf("[backend-deposit] find user %d failed: %v", id, err)
		return nil, apierrors.ErrInternal
	}

	availableBalance, frozenBalance, err := s.walletRepo.Deposit(ctx, id, currency, amount)
	if err != nil {
		logx.WithContext(ctx).Errorf("[backend-deposit] deposit for user %d currency %s failed: %v", id, currency, err)
		return nil, apierrors.ErrInternal
	}

	// 稽核記錄：legacy 完全沒有留下操作人，事後無法追查是誰手動加值。
	logx.WithContext(ctx).Infof("[backend-deposit] admin=%d user=%d currency=%s amount=%s", adminUID, id, currency, amount)

	return &backend_interface.BackendDepositResponse{
		Currency:         currency,
		AvailableBalance: availableBalance,
		FrozenBalance:    frozenBalance,
	}, nil
}

// currencyExists 檢查幣別是否存在且仍開放（currencies.is_active）。
// 幣別清單是小而穩定的參照資料，直接以既有連線查詢，不另開 repository。
func (s *backendAdminService) currencyExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := s.db.QueryRowCtx(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM currencies WHERE code = $1 AND is_active = TRUE)`,
		code,
	)
	if err != nil {
		return false, err
	}
	return exists, nil
}
