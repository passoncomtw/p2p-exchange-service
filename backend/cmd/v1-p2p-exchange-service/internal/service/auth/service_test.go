package auth_service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest/handler"
	"golang.org/x/crypto/bcrypt"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

const (
	testSecret = "test-app-access-secret-do-not-reuse"
	testExpire = int64(3600)
)

// stubUserRepo 只實作註冊會用到的方法；其餘方法沿用內嵌的 nil 介面，
// 走到未預期的路徑會直接 panic，不會靜默通過。
type stubUserRepo struct {
	userrepo.UserRepository

	existing     *entity.AppUser
	findErr      error
	createErr    error
	createCalls  int
	gotUsername  string
	gotPassHash  string
	nextID       int64
	pushTokenSet string
}

func (r *stubUserRepo) FindByUsername(_ context.Context, username string) (*entity.AppUser, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.existing != nil && r.existing.Username == username {
		return r.existing, nil
	}
	return nil, sqlx.ErrNotFound
}

func (r *stubUserRepo) Create(_ context.Context, username, passwordHash string) (*entity.AppUser, error) {
	r.createCalls++
	r.gotUsername = username
	r.gotPassHash = passwordHash
	if r.createErr != nil {
		return nil, r.createErr
	}
	id := r.nextID
	if id == 0 {
		id = 1001
	}
	return &entity.AppUser{ID: id, Username: username, PasswordHash: passwordHash}, nil
}

func (r *stubUserRepo) UpdatePushToken(_ context.Context, _ int64, token string) error {
	r.pushTokenSet = token
	return nil
}

type stubWalletRepo struct {
	walletrepo.WalletRepository

	calls       int
	gotUserID   int64
	gotCurrency string
	err         error
}

func (w *stubWalletRepo) CreateEmpty(_ context.Context, userID int64, currency string) error {
	w.calls++
	w.gotUserID = userID
	w.gotCurrency = currency
	return w.err
}

func newTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.App.AccessSecret = testSecret
	cfg.App.AccessExpire = testExpire
	return cfg
}

func newTestService(users *stubUserRepo, wallets *stubWalletRepo) AuthService {
	return New(newTestConfig(), users, wallets)
}

func assertAppErrCode(t *testing.T, err error, want int) {
	t.Helper()
	var appErr *apierrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apierrors.AppError, got %T (%v)", err, err)
	}
	if appErr.Code != want {
		t.Fatalf("expected code %d, got %d (%s)", want, appErr.Code, appErr.Message)
	}
}

// TestRegisterRejectsInvalidInput 帳號/密碼格式的邊界必須在建立任何資料之前就被擋下。
func TestRegisterRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "帳號 2 字元", username: "ab", password: "password1"},
		{name: "帳號 51 字元", username: strings.Repeat("a", 51), password: "password1"},
		{name: "帳號空字串", username: "", password: "password1"},
		{name: "帳號僅空白", username: "      ", password: "password1"},
		{name: "帳號去空白後 2 字元", username: "  ab  ", password: "password1"},
		{name: "密碼 7 字元", username: "alice", password: "1234567"},
		{name: "密碼空字串", username: "alice", password: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &stubUserRepo{}
			wallets := &stubWalletRepo{}
			svc := newTestService(users, wallets)

			out, err := svc.Register(context.Background(), tt.username, tt.password)
			if err == nil {
				t.Fatalf("expected rejection, got %+v", out)
			}
			assertAppErrCode(t, err, 400)
			if users.createCalls != 0 {
				t.Fatalf("驗證失敗仍呼叫了 Create（%d 次）", users.createCalls)
			}
			if wallets.calls != 0 {
				t.Fatalf("驗證失敗仍建立了錢包（%d 次）", wallets.calls)
			}
		})
	}
}

// TestRegisterAcceptsBoundaryInput 剛好落在合法邊界上的輸入必須通過。
func TestRegisterAcceptsBoundaryInput(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "帳號 3 字元", username: "abc", password: "12345678"},
		{name: "帳號 50 字元", username: strings.Repeat("a", 50), password: "12345678"},
		{name: "密碼 8 字元", username: "alice", password: "12345678"},
		{name: "帳號去空白後 3 字元", username: "  abc  ", password: "12345678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &stubUserRepo{}
			wallets := &stubWalletRepo{}
			svc := newTestService(users, wallets)

			out, err := svc.Register(context.Background(), tt.username, tt.password)
			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if want := strings.TrimSpace(tt.username); users.gotUsername != want {
				t.Fatalf("寫入的帳號應為去空白後的 %q，實際為 %q", want, users.gotUsername)
			}
			if out.Username != users.gotUsername {
				t.Fatalf("回應帳號 %q 與寫入帳號 %q 不一致", out.Username, users.gotUsername)
			}
		})
	}
}

// TestRegisterRejectsExistingUsername 帳號已存在時回傳 409，且不得覆寫既有帳號的密碼。
//
// 攻擊情境：對既有帳號重送註冊請求，如果流程沒有攔截而是走到 Create/UPSERT，
// 等於任何人都能用註冊 API 重設別人的密碼並直接拿到該帳號的 token。
func TestRegisterRejectsExistingUsername(t *testing.T) {
	users := &stubUserRepo{existing: &entity.AppUser{ID: 7, Username: "alice", PasswordHash: "$2a$10$existinghash"}}
	wallets := &stubWalletRepo{}
	svc := newTestService(users, wallets)

	out, err := svc.Register(context.Background(), "alice", "attacker-password")
	if err == nil {
		t.Fatalf("帳號已存在卻註冊成功，回傳 %+v", out)
	}
	assertAppErrCode(t, err, 409)
	if !errors.Is(err, apierrors.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
	if users.createCalls != 0 {
		t.Fatalf("帳號已存在仍呼叫了 Create（%d 次）", users.createCalls)
	}
	if wallets.calls != 0 {
		t.Fatalf("帳號已存在仍建立了錢包（%d 次）", wallets.calls)
	}
}

// TestRegisterExistingUsernameMessageHasNoInternalDetail 409 訊息只說明帳號已存在，不夾帶內部細節。
func TestRegisterExistingUsernameMessageHasNoInternalDetail(t *testing.T) {
	users := &stubUserRepo{existing: &entity.AppUser{ID: 7, Username: "alice"}}
	svc := newTestService(users, &stubWalletRepo{})

	_, err := svc.Register(context.Background(), "alice", "12345678")
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != apierrors.ErrUserAlreadyExists.Message {
		t.Fatalf("錯誤訊息應為固定字串 %q，實際為 %q", apierrors.ErrUserAlreadyExists.Message, got)
	}
}

// TestRegisterQueryFailureDoesNotCreateUser 查詢帳號時的非 ErrNotFound 錯誤不可被當成「帳號可用」。
//
// 攻擊情境：DB 短暫故障時若把任意查詢錯誤當成查無此人，就會在既有帳號上再建一筆同名紀錄
// （或讓後續流程以錯誤前提繼續），註冊檢查形同虛設。
func TestRegisterQueryFailureDoesNotCreateUser(t *testing.T) {
	dbErr := errors.New("pq: connection reset by peer on host db-prod-1")
	users := &stubUserRepo{findErr: dbErr}
	wallets := &stubWalletRepo{}
	svc := newTestService(users, wallets)

	_, err := svc.Register(context.Background(), "alice", "12345678")
	if err == nil {
		t.Fatal("查詢失敗卻註冊成功")
	}
	assertAppErrCode(t, err, 500)
	if users.createCalls != 0 {
		t.Fatalf("查詢失敗仍呼叫了 Create（%d 次）", users.createCalls)
	}
	if strings.Contains(err.Error(), "db-prod-1") || strings.Contains(err.Error(), "connection reset") {
		t.Fatalf("內部 DB 錯誤字串外洩：%q", err.Error())
	}
}

// TestRegisterCreateFailureIsGeneric 建立帳號失敗時只回泛用 500，不外露 DB 錯誤字串。
func TestRegisterCreateFailureIsGeneric(t *testing.T) {
	users := &stubUserRepo{createErr: errors.New(`pq: duplicate key value violates unique constraint "app_users_username_key"`)}
	svc := newTestService(users, &stubWalletRepo{})

	_, err := svc.Register(context.Background(), "alice", "12345678")
	assertAppErrCode(t, err, 500)
	if strings.Contains(err.Error(), "app_users_username_key") {
		t.Fatalf("內部 DB 錯誤字串外洩：%q", err.Error())
	}
}

// TestRegisterCreatesEmptyUSDTWallet 註冊成功必須為新帳號建立一筆 USDT 錢包。
func TestRegisterCreatesEmptyUSDTWallet(t *testing.T) {
	users := &stubUserRepo{nextID: 4242}
	wallets := &stubWalletRepo{}
	svc := newTestService(users, wallets)

	if _, err := svc.Register(context.Background(), "alice", "12345678"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if wallets.calls != 1 {
		t.Fatalf("expected 1 CreateEmpty call, got %d", wallets.calls)
	}
	if wallets.gotUserID != 4242 {
		t.Fatalf("錢包應掛在新建立的使用者 4242 上，實際為 %d", wallets.gotUserID)
	}
	if wallets.gotCurrency != "USDT" {
		t.Fatalf("expected USDT wallet, got %q", wallets.gotCurrency)
	}
}

// TestRegisterFailsWhenWalletCreationFails 錢包建立失敗即整個註冊失敗，不得回傳可用 token。
func TestRegisterFailsWhenWalletCreationFails(t *testing.T) {
	users := &stubUserRepo{}
	wallets := &stubWalletRepo{err: errors.New("pq: insert or update on table \"wallets\" violates foreign key constraint")}
	svc := newTestService(users, wallets)

	out, err := svc.Register(context.Background(), "alice", "12345678")
	if err == nil {
		t.Fatalf("錢包建立失敗卻回傳成功：%+v", out)
	}
	assertAppErrCode(t, err, 500)
	if strings.Contains(err.Error(), "foreign key constraint") {
		t.Fatalf("內部 DB 錯誤字串外洩：%q", err.Error())
	}
}

// TestRegisterStoresBcryptHashNotPlaintext 寫入 DB 的必須是 bcrypt 雜湊，絕不能是明文密碼。
func TestRegisterStoresBcryptHashNotPlaintext(t *testing.T) {
	const password = "sup3r-secret-pw"
	users := &stubUserRepo{}
	svc := newTestService(users, &stubWalletRepo{})

	if _, err := svc.Register(context.Background(), "alice", password); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if users.gotPassHash == password {
		t.Fatal("密碼以明文寫入 DB")
	}
	if !strings.HasPrefix(users.gotPassHash, "$2a$") {
		t.Fatalf("expected bcrypt hash, got %q", users.gotPassHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users.gotPassHash), []byte(password)); err != nil {
		t.Fatalf("寫入的雜湊無法比對回原密碼：%v", err)
	}
	// 相同密碼兩次註冊必須產生不同雜湊（bcrypt 每次帶不同 salt）。
	users2 := &stubUserRepo{}
	svc2 := newTestService(users2, &stubWalletRepo{})
	if _, err := svc2.Register(context.Background(), "bob", password); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if users2.gotPassHash == users.gotPassHash {
		t.Fatal("相同密碼產生了相同雜湊（salt 未生效）")
	}
}

// parseAppToken 以 App JWT 密鑰解析 token 並回傳 claims。
func parseAppToken(t *testing.T, tokenStr string) jwt.MapClaims {
	t.Helper()
	claims := jwt.MapClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("token 無法以 App 密鑰解析：%v", err)
	}
	if !tok.Valid {
		t.Fatal("token 無效")
	}
	return claims
}

// TestRegisterIssuesUsableAppToken 註冊回傳的 token 必須帶齊 App JWT 中介層需要的 claims。
func TestRegisterIssuesUsableAppToken(t *testing.T) {
	users := &stubUserRepo{nextID: 8899}
	svc := newTestService(users, &stubWalletRepo{})

	out, err := svc.Register(context.Background(), "  alice  ", "12345678")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if out.ExpireIn != testExpire {
		t.Fatalf("expected expireIn %d, got %d", testExpire, out.ExpireIn)
	}
	if out.UserID != 8899 || out.Username != "alice" {
		t.Fatalf("unexpected output: %+v", out)
	}

	claims := parseAppToken(t, out.AccessToken)
	if uid, ok := claims["uid"].(float64); !ok || int64(uid) != 8899 {
		t.Fatalf("uid claim 錯誤：%v", claims["uid"])
	}
	if username, ok := claims["username"].(string); !ok || username != "alice" {
		t.Fatalf("username claim 錯誤：%v", claims["username"])
	}
	if platform, ok := claims["platform"].(string); !ok || platform != "app" {
		t.Fatalf("platform claim 錯誤：%v", claims["platform"])
	}
	iat, _ := claims["iat"].(float64)
	exp, _ := claims["exp"].(float64)
	if int64(exp)-int64(iat) != testExpire {
		t.Fatalf("exp - iat 應為 %d，實際為 %d", testExpire, int64(exp)-int64(iat))
	}
}

// TestRegisterTokenIsRejectedByWrongSecret 註冊 token 必須以 App.AccessSecret 簽章，別的密鑰不得驗過。
func TestRegisterTokenIsRejectedByWrongSecret(t *testing.T) {
	svc := newTestService(&stubUserRepo{}, &stubWalletRepo{})
	out, err := svc.Register(context.Background(), "alice", "12345678")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	_, err = jwt.Parse(out.AccessToken, func(*jwt.Token) (any, error) {
		return []byte("some-other-secret"), nil
	})
	if err == nil {
		t.Fatal("token 竟能以其他密鑰驗證通過")
	}
}

// TestRegisterTokenPassesRealJWTMiddleware 註冊拿到的 token 必須能通過既有的 App JWT 中介層
// （go-zero rest.WithJwt 底層即 handler.Authorize），且 uid / username claim 會落進 request context，
// 後續所有 App API 都靠這個 uid 判定身分。
func TestRegisterTokenPassesRealJWTMiddleware(t *testing.T) {
	users := &stubUserRepo{nextID: 4242}
	svc := newTestService(users, &stubWalletRepo{})
	out, err := svc.Register(context.Background(), "alice", "12345678")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	var gotUID int64
	var gotUsername string
	protected := handler.Authorize(testSecret)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		// 與 handlers.ctxUID 相同的讀法：go-zero 會把數值 claim 放成 json.Number。
		if num, ok := r.Context().Value("uid").(json.Number); ok {
			gotUID, _ = num.Int64()
		}
		// go-zero 把每個非標準 claim 直接放進 context（不是包在 payload map 裡）。
		gotUsername, _ = r.Context().Value("username").(string)
	}))

	req := httptest.NewRequest(http.MethodGet, "/app/profile", nil)
	req.Header.Set("Authorization", "Bearer "+out.AccessToken)
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("註冊 token 未通過 JWT 中介層：%d %s", rec.Code, rec.Body.String())
	}
	if gotUID != 4242 {
		t.Fatalf("中介層解出的 uid 應為 4242，實際為 %d", gotUID)
	}
	if gotUsername != "alice" {
		t.Fatalf("中介層解出的 username 應為 alice，實際為 %q", gotUsername)
	}
}

// TestLoginBehaviourUnchangedAfterHelperExtraction issueAppToken 抽出後，
// Login 的成功路徑與失敗路徑行為必須與抽取前完全一致。
func TestLoginBehaviourUnchangedAfterHelperExtraction(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	users := &stubUserRepo{existing: &entity.AppUser{ID: 55, Username: "alice", PasswordHash: string(hash)}}
	svc := newTestService(users, &stubWalletRepo{})

	t.Run("成功登入的 claims 與回應不變", func(t *testing.T) {
		out, err := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "correct-password"})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if out.UserID != 55 || out.Username != "alice" || out.ExpireIn != testExpire {
			t.Fatalf("unexpected output: %+v", out)
		}
		claims := parseAppToken(t, out.AccessToken)
		if len(claims) != 5 {
			t.Fatalf("claims 數量應為 5（uid/username/platform/iat/exp），實際為 %d：%v", len(claims), claims)
		}
		if uid, _ := claims["uid"].(float64); int64(uid) != 55 {
			t.Fatalf("uid claim 錯誤：%v", claims["uid"])
		}
		if claims["username"] != "alice" || claims["platform"] != "app" {
			t.Fatalf("claims 錯誤：%v", claims)
		}
	})

	t.Run("密碼錯誤仍回模糊訊息", func(t *testing.T) {
		_, err := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "wrong-password"})
		if err == nil || err.Error() != "帳號或密碼錯誤" {
			t.Fatalf("expected 帳號或密碼錯誤, got %v", err)
		}
	})

	t.Run("帳號不存在與密碼錯誤的訊息相同", func(t *testing.T) {
		_, errNoUser := svc.Login(context.Background(), LoginInput{Username: "nobody", Password: "correct-password"})
		_, errBadPass := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "wrong-password"})
		if errNoUser == nil || errBadPass == nil {
			t.Fatal("expected both to fail")
		}
		if errNoUser.Error() != errBadPass.Error() {
			t.Fatalf("登入錯誤訊息不一致，會造成帳號列舉：%q vs %q", errNoUser.Error(), errBadPass.Error())
		}
	})
}

// TestLoginAndRegisterTokensShareSameShape 兩條路徑必須共用同一套簽發邏輯，claims 結構不得分岔。
func TestLoginAndRegisterTokensShareSameShape(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	loginUsers := &stubUserRepo{existing: &entity.AppUser{ID: 55, Username: "alice", PasswordHash: string(hash)}}
	loginSvc := newTestService(loginUsers, &stubWalletRepo{})
	loginOut, err := loginSvc.Login(context.Background(), LoginInput{Username: "alice", Password: "correct-password"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	registerUsers := &stubUserRepo{nextID: 55}
	registerSvc := newTestService(registerUsers, &stubWalletRepo{})
	registerOut, err := registerSvc.Register(context.Background(), "alice", "correct-password")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	loginClaims := parseAppToken(t, loginOut.AccessToken)
	registerClaims := parseAppToken(t, registerOut.AccessToken)
	if len(loginClaims) != len(registerClaims) {
		t.Fatalf("claims 數量不一致：login %v vs register %v", loginClaims, registerClaims)
	}
	for _, key := range []string{"uid", "username", "platform"} {
		if loginClaims[key] != registerClaims[key] {
			t.Fatalf("claim %s 不一致：login %v vs register %v", key, loginClaims[key], registerClaims[key])
		}
	}
}
