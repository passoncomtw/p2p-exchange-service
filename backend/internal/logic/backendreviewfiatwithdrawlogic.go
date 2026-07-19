package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendReviewFiatWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendReviewFiatWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendReviewFiatWithdrawLogic {
	return &BackendReviewFiatWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BackendReviewFiatWithdrawLogic) Review(reviewerID int64, req *types.BackendReviewFiatWithdrawalRequest) error {
	if req.Action != "approve" && req.Action != "reject" {
		return apperrors.New(400, "action 必須為 approve 或 reject")
	}
	if req.Action == "reject" && req.Reason == "" {
		return apperrors.New(400, "拒絕時必須填寫原因")
	}

	w, err := l.svcCtx.FiatWithdraw.FindByID(l.ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apperrors.ErrNotFound
		}
		return err
	}
	if w.Status != "pending" {
		return apperrors.New(400, "此申請已審核完畢")
	}

	now := time.Now()

	if req.Action == "approve" {
		// Deduct from frozen balance → permanent deduction, ledger type fiat_withdraw
		if err := l.svcCtx.Wallet.DeductFrozenBalance(l.ctx, w.UserID, w.Currency, w.Amount, "fiat_withdraw"); err != nil {
			return err
		}
		if err := l.svcCtx.FiatWithdraw.UpdateApproved(l.ctx, w.ID, reviewerID, now); err != nil {
			l.Errorf("[backend-review] UpdateApproved id=%d: %v", w.ID, err)
			return err
		}
		l.Infof("[backend-review] approved fiat_withdrawal id=%d user=%d amount=%s reviewer=%d", w.ID, w.UserID, w.Amount, reviewerID)
	} else {
		// Reject: return frozen balance to available
		if err := l.svcCtx.Wallet.UnfreezeBalance(l.ctx, w.UserID, w.Currency, w.Amount); err != nil {
			return err
		}
		if err := l.svcCtx.FiatWithdraw.UpdateRejected(l.ctx, w.ID, reviewerID, now, req.Reason); err != nil {
			l.Errorf("[backend-review] UpdateRejected id=%d: %v", w.ID, err)
			return err
		}
		l.Infof("[backend-review] rejected fiat_withdrawal id=%d user=%d reason=%q reviewer=%d", w.ID, w.UserID, req.Reason, reviewerID)
	}
	return nil
}
