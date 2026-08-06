package backend_auth_service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	backenduserrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/backend_user"
)

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	Token     string
	ExpiresIn int64
}

type BackendAuthService interface {
	Login(ctx context.Context, req LoginInput) (*LoginOutput, error)
}

type backendAuthService struct {
	cfg             *config.Config
	backendUserRepo backenduserrepo.BackendUserRepository
}

func New(cfg *config.Config, backendUserRepo backenduserrepo.BackendUserRepository) BackendAuthService {
	return &backendAuthService{cfg: cfg, backendUserRepo: backendUserRepo}
}

func (s *backendAuthService) Login(ctx context.Context, req LoginInput) (*LoginOutput, error) {
	user, err := s.backendUserRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"uid":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"platform": "backend",
		"iat":      now.Unix(),
		"exp":      now.Unix() + s.cfg.Backend.AccessExpire,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.Backend.AccessSecret))
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:     tokenStr,
		ExpiresIn: s.cfg.Backend.AccessExpire,
	}, nil
}
