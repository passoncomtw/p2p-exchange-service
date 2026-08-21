package config

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "internal/constants/config.yaml", "the config file")

type Config struct {
	rest.RestConf
	// App JWT 設定（AccessSecret 敏感，透過環境變數注入，與 legacy 共用同一組密鑰）
	App struct {
		AccessSecret string `json:",env=APP_JWT_ACCESS_SECRET"`
		AccessExpire int64
	}
	// Backend JWT 設定（AccessSecret 敏感，透過環境變數注入，與 legacy 共用同一組密鑰）
	Backend struct {
		AccessSecret string `json:",env=BACKEND_JWT_ACCESS_SECRET"`
		AccessExpire int64
	}
	// Database 連線設定（DSN 含帳密，透過環境變數注入，不得 commit，與 legacy 共用同一個 DB）
	Database struct {
		DSN string `json:"dsn,env=DATABASE_DSN"`
	} `json:"database"`
	// Redis 連線設定（Addr/Password 敏感，透過環境變數注入，與 legacy 共用同一個 Redis）
	Redis struct {
		Addr     string `json:"addr,env=REDIS_ADDR"`
		Password string `json:"password,optional,env=REDIS_PASSWORD"`
		PoolSize int    `json:"poolSize"`
	} `json:"redis"`
	WebSocket struct {
		Enabled bool   `json:"enabled"`
		Host    string `json:"host"`
		Port    int    `json:"port"`
	} `json:"webSocket"`
	Nats struct {
		URL          string
		CredsPath    string
		User         string
		Password     string
		StreamName   string
		ConsumerName string
	}
}

// Validate 檢查敏感欄位是否已透過環境變數注入；config.yaml 本身不再提供實際值，
// 缺任何一項就代表對應的環境變數沒設，寧可啟動失敗也不要用空字串連線。
func (c Config) Validate() error {
	switch {
	case c.App.AccessSecret == "":
		return fmt.Errorf("config: App.AccessSecret is empty — set APP_JWT_ACCESS_SECRET")
	case c.Backend.AccessSecret == "":
		return fmt.Errorf("config: Backend.AccessSecret is empty — set BACKEND_JWT_ACCESS_SECRET")
	case c.Database.DSN == "":
		return fmt.Errorf("config: Database.DSN is empty — set DATABASE_DSN")
	case c.Redis.Addr == "":
		return fmt.Errorf("config: Redis.Addr is empty — set REDIS_ADDR")
	}
	return nil
}

func NewConfig() *Config {
	var config Config
	conf.MustLoad(*configFile, &config) // conf.MustLoad 會自動呼叫 Config.Validate()
	return &config
}
