package logic

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendMemberDepositLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendMemberDepositLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendMemberDepositLogic {
	return &BackendMemberDepositLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BackendMemberDepositLogic) Deposit(req *types.BackendDepositRequest) (*types.BackendDepositResponse, error) {
	amt, ok := new(big.Float).SetString(req.Amount)
	if !ok || amt.Sign() <= 0 {
		return nil, apperrors.New(400, "amount must be a positive number")
	}

	if req.Currency == "" {
		return nil, apperrors.New(400, "currency is required")
	}

	exists, err := l.svcCtx.Currency.Exists(l.ctx, req.Currency)
	if err != nil {
		l.Errorf("deposit: check currency %s failed: %v", req.Currency, err)
		return nil, apperrors.ErrInternal
	}
	if !exists {
		return nil, apperrors.New(400, "unsupported currency")
	}

	_, err = l.svcCtx.AppUser.FindByID(l.ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apperrors.ErrNotFound
		}
		l.Errorf("deposit: find user %d failed: %v", req.ID, err)
		return nil, apperrors.ErrInternal
	}

	wallet, err := l.svcCtx.Wallet.Deposit(l.ctx, req.ID, req.Currency, req.Amount)
	if err != nil {
		l.Errorf("deposit: wallet deposit for user %d currency %s failed: %v", req.ID, req.Currency, err)
		return nil, apperrors.ErrInternal
	}

	return &types.BackendDepositResponse{
		Currency:         wallet.Currency,
		AvailableBalance: wallet.AvailableBalance,
		FrozenBalance:    wallet.FrozenBalance,
	}, nil
}
