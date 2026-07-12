package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppListWalletLedgersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListWalletLedgersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListWalletLedgersLogic {
	return &AppListWalletLedgersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListWalletLedgersLogic) ListWalletLedgers(userID int64, req *types.ListWalletLedgersRequest) (*types.ListWalletLedgersResponse, error) {
	wallet, err := l.svcCtx.Wallet.FindOne(l.ctx, userID, req.Currency)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	total, err := l.svcCtx.WalletLedger.CountByWalletID(l.ctx, wallet.ID)
	if err != nil {
		return nil, err
	}

	ledgers, err := l.svcCtx.WalletLedger.ListByWalletID(l.ctx, wallet.ID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	list := make([]types.WalletLedgerItem, 0, len(ledgers))
	for _, le := range ledgers {
		list = append(list, types.WalletLedgerItem{
			Type:         le.Type,
			Amount:       le.Amount,
			BalanceAfter: le.BalanceAfter,
			RefOrderNo:   le.RefOrderNo,
			CreatedAt:    le.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return &types.ListWalletLedgersResponse{
		List:  list,
		Total: total,
	}, nil
}
