package logic

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppLoginLogic {
	return &AppLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppLoginLogic) Login(req *types.LoginRequest) (*types.AppLoginResponse, error) {
	user, err := l.svcCtx.AppUser.FindByUsername(l.ctx, req.Username)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrInvalidCredentials
		}
		l.Errorf("app: query user %s failed: %v", req.Username, err)
		return nil, apierrors.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apierrors.ErrInvalidCredentials
	}

	token, err := l.generateToken(user.ID, user.Username)
	if err != nil {
		l.Errorf("app: failed to generate token for user %s: %v", user.Username, err)
		return nil, apierrors.ErrInternal
	}

	return &types.AppLoginResponse{
		AccessToken: token,
		ExpireIn:    l.svcCtx.Config.App.AccessExpire,
		User: types.AppLoginUserInfo{
			ID:      user.ID,
			Account: user.Username,
			Name:    user.Username, // 第一階段以 username 作為 name
		},
	}, nil
}

func (l *AppLoginLogic) generateToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"uid":      userID,
		"username": username,
		"platform": "app",
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(l.svcCtx.Config.App.AccessExpire) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(l.svcCtx.Config.App.AccessSecret))
}
