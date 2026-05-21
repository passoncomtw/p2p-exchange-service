package logic

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	if err := l.verifyCredentials(req.Username, req.Password); err != nil {
		return nil, err
	}

	token, err := l.generateToken(req.Username)
	if err != nil {
		l.Errorf("failed to generate token for user %s: %v", req.Username, err)
		return nil, errors.New("failed to generate token")
	}

	return &types.LoginResponse{
		Token:     token,
		ExpiresIn: l.svcCtx.Config.Auth.ExpireIn,
	}, nil
}

func (l *LoginLogic) verifyCredentials(username, password string) error {
	// TODO: 替換為實際的資料庫查詢與密碼雜湊驗證
	if username == "admin" && password == "password" {
		return nil
	}
	return ErrInvalidCredentials
}

func (l *LoginLogic) generateToken(username string) (string, error) {
	expireIn := l.svcCtx.Config.Auth.ExpireIn
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": username,
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(expireIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(l.svcCtx.Config.Auth.Secret))
}
