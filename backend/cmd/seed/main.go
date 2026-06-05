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
var appUsers = []struct {
	Username string
	Password string
}{
	{"testdemo001", "a12345678"},
	{"testdemo002", "a12345678"},
	{"testdemo003", "a12345678"},
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

	fmt.Println("seed completed")
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
