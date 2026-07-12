package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppListWalletsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListWalletsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListWalletsLogic {
	return &AppListWalletsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListWalletsLogic) ListWallets(userID int64) (*types.ListWalletsResponse, error) {
	wallets, err := l.svcCtx.Wallet.FindByUserID(l.ctx, userID)
	if err != nil {
		return nil, err
	}

	list := make([]types.WalletItem, 0, len(wallets))
	for _, w := range wallets {
		list = append(list, types.WalletItem{
			Currency:         w.Currency,
			AvailableBalance: w.AvailableBalance,
			FrozenBalance:    w.FrozenBalance,
		})
	}

	return &types.ListWalletsResponse{List: list}, nil
}
