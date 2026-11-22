package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// TestSensitiveFieldsRequireEnv 確保敏感欄位為空時 Config.Validate() 會失敗，避免用空字串靜默連線；
// 設定對應環境變數後，conf.Load 應成功且欄位值來自環境變數。
//
// 未設環境變數的階段直接呼叫 Validate()，不經過 conf.Load：
// go-zero 的 proc.Env 會快取第一次讀到的值（含空字串），先以空環境 Load 一次會讓後續 t.Setenv 失效。
func TestSensitiveFieldsRequireEnv(t *testing.T) {
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("expected Validate to fail when sensitive fields are empty, got nil")
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
