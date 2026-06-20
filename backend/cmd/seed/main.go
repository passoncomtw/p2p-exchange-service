// cmd/seed/main.go — 一次性執行，寫入開發測試帳號
// 執行：go run cmd/seed/main.go -f etc/config.yaml
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
	"p2p-exchange/internal/config"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

// App 使用者（一般會員）
// demo_user：v1 無登入時固定使用的建立者。
// trader_*：seed 樣本掛單的他人帳號（createdBy != demo_user）。
var appUsers = []struct {
	Username string
	Password string
}{
	{"testdemo001", "a12345678"},
	{"testdemo002", "a12345678"},
	{"testdemo003", "a12345678"},
	{"demo_user", "a12345678"},
	{"trader_alice", "a12345678"},
	{"trader_bob", "a12345678"},
	{"trader_carol", "a12345678"},
}

// v1 樣本掛單（沿用 listings 表）。狀態以 DB 詞彙表示：active = open。
var sampleListings = []struct {
	Username      string
	Type          string
	Price         float64
	Quantity      float64
	PaymentMethod string
	Status        string
}{
	{"trader_alice", "sell", 32.5, 100, "bank_transfer", "active"},
	{"trader_bob", "buy", 31.8, 200, "convenience_store", "active"},
	{"trader_carol", "sell", 32.1, 50, "bank_transfer", "completed"},
	{"trader_alice", "buy", 31.5, 300, "bank_transfer", "cancelled"},
}

// 後台管理使用者
var backendUsers = []struct {
	Username string
	Password string
	Role     string
}{
	{"admin001", "admin@1234", "admin"},
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	conn := sqlx.NewSqlConn("pgx", c.Database.DSN)
	ctx := context.Background()

	for _, u := range appUsers {
		if err := seedAppUser(ctx, conn, u.Username, u.Password); err != nil {
			log.Fatalf("seed app_user %s failed: %v", u.Username, err)
		}
	}

	for _, u := range backendUsers {
		if err := seedBackendUser(ctx, conn, u.Username, u.Password, u.Role); err != nil {
			log.Fatalf("seed backend_user %s failed: %v", u.Username, err)
		}
	}

	for _, s := range sampleListings {
		if err := seedSampleListing(ctx, conn, s.Username, s.Type, s.Price, s.Quantity, s.PaymentMethod, s.Status); err != nil {
			log.Fatalf("seed sample listing for %s failed: %v", s.Username, err)
		}
	}

	fmt.Println("seed completed")
}

// seedSampleListing 寫入一筆他人樣本掛單；以 (user, type, price, quantity, status) 去重避免重複植入。
func seedSampleListing(ctx context.Context, conn sqlx.SqlConn, username, listType string, price, quantity float64, paymentMethod, status string) error {
	var userID int64
	if err := conn.QueryRowCtx(ctx, &userID,
		`SELECT id FROM app_users WHERE username = $1`, username,
	); err != nil {
		return err
	}

	var existing int64
	if err := conn.QueryRowCtx(ctx, &existing,
		`SELECT COUNT(*) FROM listings
		 WHERE user_id = $1 AND type = $2 AND price = $3 AND total_amount = $4 AND status = $5`,
		userID, listType, price, quantity, status,
	); err != nil {
		return err
	}
	if existing > 0 {
		fmt.Printf("sample listing already exists: %s %s %.2f x %.2f (%s)\n", username, listType, price, quantity, status)
		return nil
	}

	maxOrderFiat := price * quantity
	_, err := conn.ExecCtx(ctx,
		`INSERT INTO listings (user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, payment_method_label, status, created_at, updated_at)
		 VALUES ($1, $2, 'USDT', 'TWD', $3, $3, $4, 0, $5, $6, $7, NOW(), NOW())`,
		userID, listType, quantity, price, maxOrderFiat, paymentMethod, status,
	)
	if err != nil {
		return err
	}
	fmt.Printf("sample listing inserted: %s %s %.2f x %.2f (%s)\n", username, listType, price, quantity, status)
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func seedAppUser(ctx context.Context, conn sqlx.SqlConn, username, password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = conn.ExecCtx(ctx,
		`INSERT INTO app_users (username, password_hash)
		 VALUES ($1, $2)
		 ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash`,
		username, hash,
	)
	if err != nil {
		return err
	}
	fmt.Printf("app_user upserted: %s\n", username)
	return nil
}

func seedBackendUser(ctx context.Context, conn sqlx.SqlConn, username, password, role string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = conn.ExecCtx(ctx,
		`INSERT INTO backend_users (username, password_hash, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role`,
		username, hash, role,
	)
	if err != nil {
		return err
	}
	fmt.Printf("backend_user upserted: %s\n", username)
	return nil
}
