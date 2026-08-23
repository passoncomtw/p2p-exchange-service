package backend_admin_service

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	apierrors "p2p-exchange/internal/errors"
)

const (
	// fiatWithdrawLedgerType 核可提領時寫入 wallet_ledgers 的帳本類型。
	fiatWithdrawLedgerType = "fiat_withdraw"

	reviewActionApprove = "approve"
	reviewActionReject  = "reject"

	// defaultFiatWithdrawPageSize limit 未帶或非正數時的預設筆數（與 legacy 相同）。
	defaultFiatWithdrawPageSize = 20
)

// ListFiatWithdrawals 後台分頁查詢法幣提領申請。
// 回傳的銀行帳號為完整值（審核人員需核對匯款資訊），此路由由 Backend JWT 保護。
func (s *backendAdminService) ListFiatWithdrawals(ctx context.Context, status string, limit, offset int64) (*backend_interface.BackendListFiatWithdrawalsResponse, error) {
	// handler 的 default tag 只在參數缺席時生效，明確帶入的負數仍會進到這裡；
	// 負數會讓 SQL 的 LIMIT / OFFSET 直接報錯，因此在此收斂成合法值。
	if limit <= 0 {
		limit = defaultFiatWithdrawPageSize
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.fiatWithdrawRepo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		logx.WithContext(ctx).Errorf("[backend-fiat-withdraw] list status=%s failed: %v", status, err)
		return nil, apierrors.ErrInternal
	}

	total, err := s.fiatWithdrawRepo.CountByStatus(ctx, status)
	if err != nil {
		logx.WithContext(ctx).Errorf("[backend-fiat-withdraw] count status=%s failed: %v", status, err)
		return nil, apierrors.ErrInternal
	}

	items := make([]backend_interface.FiatWithdrawalItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, backend_interface.FiatWithdrawalItem{
			ID:           r.ID,
			UserID:       r.UserID,
			Currency:     r.Currency,
			Amount:       r.Amount,
			BankCode:     r.BankCode,
			BankAccount:  r.BankAccount,
			AccountName:  r.AccountName,
			Status:       r.Status,
			ReviewedBy:   r.ReviewedBy,
			RejectReason: r.RejectReason,
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		})
	}

	return &backend_interface.BackendListFiatWithdrawalsResponse{List: items, Total: total}, nil
}

// ReviewFiatWithdrawal 審核法幣提領申請。
//
// reviewerID 為審核人（後台帳號 ID），由 handler 從 Backend JWT context 取得；
// 本方法不接受任何來自 request body 的審核人資訊。
//
// 併發安全（legacy P0-1 的修正）：狀態轉換與資金異動收在同一個交易內，且**順序固定為
// 先狀態轉換、確認 RowsAffected = 1、再動資金**。兩個並發審核只有一個能通過
// `WHERE status = 'pending'` 的第一道 UPDATE，另一個在狀態轉換就被擋下（0 列）而完全不碰資金；
// 資金異動若失敗（例如凍結餘額不足），整筆交易連同狀態轉換一起回滾，
// 不會留下「狀態已改但錢沒動」的中間態。
//
// 已被審核過的申請回傳 409：審核是人為決策，不可像自動化流程那樣把重複觸發當成冪等成功，
// 否則後台人員會誤以為這次操作生效。
func (s *backendAdminService) ReviewFiatWithdrawal(ctx context.Context, reviewerID, id int64, action, reason string) error {
	if action != reviewActionApprove && action != reviewActionReject {
		return apierrors.New(400, "action 必須為 approve 或 reject")
	}
	if action == reviewActionReject && reason == "" {
		return apierrors.New(400, "拒絕時必須填寫原因")
	}

	// 這次讀取只用來取得金額與幣別等不可變欄位；w.Status 刻意不拿來當審核前置檢查，
	// 是否能審核一律以下方條件式 UPDATE 的 RowsAffected 為準。
	w, err := s.fiatWithdrawRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return apierrors.ErrNotFound
		}
		logx.WithContext(ctx).Errorf("[backend-fiat-withdraw] find %d failed: %v", id, err)
		return apierrors.ErrInternal
	}

	now := time.Now()
	err = s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var affected int64
		var err error
		if action == reviewActionApprove {
			affected, err = s.fiatWithdrawRepo.UpdateApprovedInTx(ctx, session, id, reviewerID, now)
		} else {
			affected, err = s.fiatWithdrawRepo.UpdateRejectedInTx(ctx, session, id, reviewerID, now, reason)
		}
		if err != nil {
			return err
		}
		if affected == 0 {
			// 狀態轉換就先被擋下，完全不會執行到下面的資金異動。
			return apierrors.New(409, "此申請已審核完畢")
		}

		if action == reviewActionApprove {
			// 核可：自凍結餘額永久扣除（款項改由線下匯款撥付）。
			return s.walletRepo.DeductFrozenBalanceWithLedgerTypeInTx(ctx, session, w.UserID, w.Currency, w.Amount, fiatWithdrawLedgerType)
		}
		// 駁回：凍結金額退回可用餘額。
		return s.walletRepo.UnfreezeBalanceInTx(ctx, session, w.UserID, w.Currency, w.Amount)
	})
	if err != nil {
		// 已具語意的錯誤（409 衝突、400 餘額不足）原樣回傳；其餘 DB 錯誤一律收斂成 500，
		// 不把底層錯誤訊息帶到 API 回應。
		var appErr *apierrors.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		logx.WithContext(ctx).Errorf("[backend-fiat-withdraw] review %d action=%s reviewer=%d failed: %v", id, action, reviewerID, err)
		return apierrors.ErrInternal
	}

	// 稽核記錄：留下是哪個後台帳號核可／駁回了哪一筆提領。
	logx.WithContext(ctx).Infof("[backend-fiat-withdraw] reviewer=%d withdrawal=%d action=%s user=%d currency=%s amount=%s",
		reviewerID, id, action, w.UserID, w.Currency, w.Amount)
	return nil
}
