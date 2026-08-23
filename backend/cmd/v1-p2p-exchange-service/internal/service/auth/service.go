package auth_service

import (
	"context"
	"errors"
	"strings"
	"time"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const (
	// usernameMinLen / usernameMaxLen 註冊帳號去除頭尾空白後的長度限制（與 legacy 一致）。
	usernameMinLen = 3
	usernameMaxLen = 50
	// passwordMinLen 註冊密碼的最短長度（與 legacy 一致）。
	passwordMinLen = 8
	// defaultWalletCurrency 註冊時預先建立的錢包幣別。
	defaultWalletCurrency = "USDT"
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
	// Register 建立 App 使用者，並比照登入成功回傳可直接使用的 access token。
	// 帳號已存在時回傳 409；格式不符回傳 400；其餘一律回傳不含內部細節的 500。
	Register(ctx context.Context, username, password string) (*LoginOutput, error)
}

type authService struct {
	cfg        *config.Config
	userRepo   userrepo.UserRepository
	walletRepo walletrepo.WalletRepository
}

func New(cfg *config.Config, userRepo userrepo.UserRepository, walletRepo walletrepo.WalletRepository) AuthService {
	return &authService{cfg: cfg, userRepo: userRepo, walletRepo: walletRepo}
}

func (s *authService) Login(ctx context.Context, req LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}

	tokenStr, err := s.issueAppToken(user.ID, user.Username)
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

func (s *authService) Register(ctx context.Context, username, password string) (*LoginOutput, error) {
	username = strings.TrimSpace(username)
	if len(username) < usernameMinLen || len(username) > usernameMaxLen {
		return nil, apierrors.New(400, "username must be between 3 and 50 characters")
	}
	if len(password) < passwordMinLen {
		return nil, apierrors.New(400, "password must be at least 8 characters")
	}

	// 查無帳號才可繼續；其餘查詢錯誤一律視為內部錯誤，不可當成「帳號可用」放行。
	_, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return nil, apierrors.ErrUserAlreadyExists
	}
	if !errors.Is(err, sqlx.ErrNotFound) {
		// 只記錄帳號與錯誤，絕不記錄密碼或雜湊值。
		logx.WithContext(ctx).Errorf("register: query user %s failed: %v", username, err)
		return nil, apierrors.ErrInternal
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logx.WithContext(ctx).Errorf("register: hash password for %s failed: %v", username, err)
		return nil, apierrors.ErrInternal
	}

	user, err := s.userRepo.Create(ctx, username, string(hash))
	if err != nil {
		logx.WithContext(ctx).Errorf("register: create user %s failed: %v", username, err)
		return nil, apierrors.ErrInternal
	}

	// 與 legacy 相同：錢包建立失敗即整個註冊失敗（不包成單一交易，維持原行為）。
	if err := s.walletRepo.CreateEmpty(ctx, user.ID, defaultWalletCurrency); err != nil {
		logx.WithContext(ctx).Errorf("register: create %s wallet for user %d failed: %v", defaultWalletCurrency, user.ID, err)
		return nil, apierrors.ErrInternal
	}

	tokenStr, err := s.issueAppToken(user.ID, user.Username)
	if err != nil {
		logx.WithContext(ctx).Errorf("register: issue token for user %d failed: %v", user.ID, err)
		return nil, apierrors.ErrInternal
	}

	return &LoginOutput{
		AccessToken: tokenStr,
		ExpireIn:    s.cfg.App.AccessExpire,
		UserID:      user.ID,
		Username:    user.Username,
	}, nil
}

// issueAppToken 簽發 App 端 JWT（HS256，密鑰為 App.AccessSecret）。
// Login 與 Register 共用同一份 claims 與過期時間計算，避免兩條路徑的 token 內容分岔。
func (s *authService) issueAppToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"uid":      userID,
		"username": username,
		"platform": "app",
		"iat":      now.Unix(),
		"exp":      now.Unix() + s.cfg.App.AccessExpire,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.App.AccessSecret))
}
