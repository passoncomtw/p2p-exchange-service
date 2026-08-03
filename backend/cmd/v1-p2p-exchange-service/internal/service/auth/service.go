package auth_service

import (
	"context"
	"errors"
	"time"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	AccessToken string
	ExpireIn    int64
	UserID      int64
	Username    string
}

type AuthService interface {
	Login(ctx context.Context, req LoginInput) (*LoginOutput, error)
}

type authService struct {
	cfg      *config.Config
	userRepo userrepo.UserRepository
}

func New(cfg *config.Config, userRepo userrepo.UserRepository) AuthService {
	return &authService{cfg: cfg, userRepo: userRepo}
}

func (s *authService) Login(ctx context.Context, req LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
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
		"platform": "app",
		"iat":      now.Unix(),
		"exp":      now.Unix() + s.cfg.App.AccessExpire,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.App.AccessSecret))
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken: tokenStr,
		ExpireIn:    s.cfg.App.AccessExpire,
		UserID:      user.ID,
		Username:    user.Username,
	}, nil
}
