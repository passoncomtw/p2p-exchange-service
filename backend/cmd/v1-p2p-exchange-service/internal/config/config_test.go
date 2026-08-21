package config

import (
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// TestSensitiveFieldsRequireEnv 確保 AccessSecret/DSN/Redis 等敏感欄位在 config.yaml 留空時，
// conf.Load 會自動呼叫 Config.Validate()（go-zero 的 Validator 慣例）並在環境變數沒設時失敗，
// 避免用空字串靜默連線；設定對應環境變數後，Load 應成功且欄位值來自環境變數。
func TestSensitiveFieldsRequireEnv(t *testing.T) {
	os.Unsetenv("APP_JWT_ACCESS_SECRET")
	os.Unsetenv("BACKEND_JWT_ACCESS_SECRET")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("REDIS_ADDR")

	var c Config
	if err := conf.Load("../constants/config.yaml", &c); err == nil {
		t.Fatal("expected conf.Load to fail when sensitive env vars are unset, got nil")
	}

	t.Setenv("APP_JWT_ACCESS_SECRET", "test-app-secret")
	t.Setenv("BACKEND_JWT_ACCESS_SECRET", "test-backend-secret")
	t.Setenv("DATABASE_DSN", "postgresql://test:test@localhost/test")
	t.Setenv("REDIS_ADDR", "localhost:6379")

	var c2 Config
	if err := conf.Load("../constants/config.yaml", &c2); err != nil {
		t.Fatalf("expected conf.Load to succeed with env vars set, got: %v", err)
	}
	if c2.App.AccessSecret != "test-app-secret" {
		t.Errorf("App.AccessSecret = %q, want %q", c2.App.AccessSecret, "test-app-secret")
	}
	if c2.Backend.AccessSecret != "test-backend-secret" {
		t.Errorf("Backend.AccessSecret = %q, want %q", c2.Backend.AccessSecret, "test-backend-secret")
	}
	if c2.Database.DSN != "postgresql://test:test@localhost/test" {
		t.Errorf("Database.DSN = %q, want the env-injected value", c2.Database.DSN)
	}
}
