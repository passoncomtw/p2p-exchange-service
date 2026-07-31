package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppGetCryptoDepositInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppGetCryptoDepositInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetCryptoDepositInfoLogic {
	return &AppGetCryptoDepositInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppGetCryptoDepositInfoLogic) GetDepositInfo(userID int64) (*types.GetCryptoDepositInfoResponse, error) {
	conf := l.svcCtx.Config.Tron
	if !conf.IsEnabled() {
		return nil, fmt.Errorf("USDT 充值功能尚未開放")
	}
	// Memo = 8-char zero-padded hex of user ID, used to identify the depositor on the hot wallet
	memo := fmt.Sprintf("%08x", userID)
	return &types.GetCryptoDepositInfoResponse{
		Address:         conf.HotWalletAddress,
		Memo:            memo,
		Network:         conf.Network,
		Currency:        "USDT",
		ContractAddress: conf.USDTContractAddress,
	}, nil
}

// ── AppListCryptoDepositsLogic ────────────────────────────────────────────────

type AppListCryptoDepositsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListCryptoDepositsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListCryptoDepositsLogic {
	return &AppListCryptoDepositsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListCryptoDepositsLogic) List(userID int64, req *types.AppListCryptoDepositsRequest) (*types.AppListCryptoDepositsResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := l.svcCtx.CryptoDeposit.ListByUserID(l.ctx, userID, limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	total, err := l.svcCtx.CryptoDeposit.CountByUserID(l.ctx, userID)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]types.CryptoDepositItem, 0, len(rows))
	for _, r := range rows {
		item := types.CryptoDepositItem{
			ID:        r.ID,
			Currency:  r.Currency,
			Amount:    r.Amount,
			TxHash:    r.TxHash,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		}
		if r.ConfirmedAt != nil {
			s := r.ConfirmedAt.Format(time.RFC3339)
			item.ConfirmedAt = &s
		}
		items = append(items, item)
	}

	return &types.AppListCryptoDepositsResponse{List: items, Total: total}, nil
}
