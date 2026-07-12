package logic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppRegisterLogic {
	return &AppRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppRegisterLogic) Register(req *types.RegisterRequest) (*types.AppLoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	if len(username) < 3 || len(username) > 50 {
		return nil, apperrors.New(400, "username must be between 3 and 50 characters")
	}
	if len(req.Password) < 8 {
		return nil, apperrors.New(400, "password must be at least 8 characters")
	}

	_, err := l.svcCtx.AppUser.FindByUsername(l.ctx, username)
	if err != nil && err != sqlx.ErrNotFound {
		l.Errorf("register: query user %s failed: %v", username, err)
		return nil, apperrors.ErrInternal
	}
	if err == nil {
		return nil, apperrors.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("register: bcrypt failed: %v", err)
		return nil, apperrors.ErrInternal
	}

	user, err := l.svcCtx.AppUser.Create(l.ctx, username, string(hash))
	if err != nil {
		l.Errorf("register: create user %s failed: %v", username, err)
		return nil, apperrors.ErrInternal
	}

	if err := l.svcCtx.Wallet.Create(l.ctx, user.ID, "USDT"); err != nil {
		l.Errorf("register: create USDT wallet for user %d failed: %v", user.ID, err)
		return nil, apperrors.ErrInternal
	}

	loginLogic := NewAppLoginLogic(l.ctx, l.svcCtx)
	token, err := loginLogic.generateToken(user.ID, user.Username)
	if err != nil {
		l.Errorf("register: generate token for user %s failed: %v", user.Username, err)
		return nil, apperrors.ErrInternal
	}

	return &types.AppLoginResponse{
		AccessToken: token,
		ExpireIn:    l.svcCtx.Config.App.AccessExpire,
		User: types.AppLoginUserInfo{
			ID:      user.ID,
			Account: user.Username,
			Name:    user.Username,
		},
	}, nil
}
