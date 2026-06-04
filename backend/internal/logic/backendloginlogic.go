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
	user, err := l.svcCtx.BackendUser.FindByUsername(l.ctx, req.Username)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrInvalidCredentials
		}
		l.Errorf("backend: query user %s failed: %v", req.Username, err)
		return nil, apierrors.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apierrors.ErrInvalidCredentials
	}

	token, err := l.generateToken(user.ID, user.Username)
	if err != nil {
		l.Errorf("backend: failed to generate token for user %s: %v", user.Username, err)
		return nil, apierrors.ErrInternal
	}

	return &types.LoginResponse{
		Token:     token,
		ExpiresIn: l.svcCtx.Config.Backend.AccessExpire,
	}, nil
}

func (l *BackendLoginLogic) generateToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      userID,    // 與 App token 一致，存 int64 userID
		"username": username,  // 獨立欄位存使用者名稱
		"platform": "backend", // 平台識別，用於二次驗證
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(l.svcCtx.Config.Backend.AccessExpire) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(l.svcCtx.Config.Backend.AccessSecret))
}
