package logic

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendLoginLogic {
	return &BackendLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BackendLoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	if err := l.verifyCredentials(req.Username, req.Password); err != nil {
		return nil, err
	}

	token, err := l.generateToken(req.Username)
	if err != nil {
		l.Errorf("backend: failed to generate token for user %s: %v", req.Username, err)
		return nil, apierrors.ErrInternal
	}

	return &types.LoginResponse{
		Token:     token,
		ExpiresIn: l.svcCtx.Config.Backend.AccessExpire,
	}, nil
}

func (l *BackendLoginLogic) verifyCredentials(username, password string) error {
	// TODO: 替換為實際的資料庫查詢與密碼雜湊驗證
	if username == "admin001" && password == "admin@1234" {
		return nil
	}
	return apierrors.ErrInvalidCredentials
}

func (l *BackendLoginLogic) generateToken(username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      username,
		"platform": "backend",
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(l.svcCtx.Config.Backend.AccessExpire) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(l.svcCtx.Config.Backend.AccessSecret))
}
