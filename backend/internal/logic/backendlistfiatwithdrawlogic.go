package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendListFiatWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendListFiatWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendListFiatWithdrawLogic {
	return &BackendListFiatWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BackendListFiatWithdrawLogic) List(req *types.BackendListFiatWithdrawalsRequest) (*types.BackendListFiatWithdrawalsResponse, error) {
	rows, err := l.svcCtx.FiatWithdraw.ListByStatus(l.ctx, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	total, err := l.svcCtx.FiatWithdraw.CountByStatus(l.ctx, req.Status)
	if err != nil {
		return nil, err
	}

	list := make([]types.FiatWithdrawalItem, 0, len(rows))
	for _, w := range rows {
		list = append(list, types.FiatWithdrawalItem{
			ID:           w.ID,
			UserID:       w.UserID,
			Currency:     w.Currency,
			Amount:       w.Amount,
			BankCode:     w.BankCode,
			BankAccount:  w.BankAccount,
			AccountName:  w.AccountName,
			Status:       w.Status,
			ReviewedBy:   w.ReviewedBy,
			RejectReason: w.RejectReason,
			CreatedAt:    w.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return &types.BackendListFiatWithdrawalsResponse{List: list, Total: total}, nil
}
