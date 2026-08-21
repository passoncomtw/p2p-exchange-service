package config

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "internal/constants/config.yaml", "the config file")

// TronConf 平台 Tron 熱錢包設定（欄位、環境變數、default 值與 legacy internal/config.TronConf 完全一致，
// 兩邊共用同一組環境變數注入）。全部欄位皆為 optional：功能可處於關閉狀態，不納入 Validate() 必填檢查。
type TronConf struct {
	Network                 string `json:",default=nile"`
	TronGridURL             string `json:",optional,env=TRON_GRID_URL"`
	TronGridAPIKey          string `json:",optional,env=TRON_GRID_API_KEY"`
	HotWalletAddress        string `json:",optional,env=TRON_HOT_WALLET_ADDRESS"`
	HotWalletPrivateKey     string `json:",optional,env=TRON_HOT_WALLET_PRIVATE_KEY"`
	USDTContractAddress     string `json:",default=TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj"`
	ConfirmationBlocks      int    `json:",default=6"`
	ScanIntervalSeconds     int    `json:",default=30"`
	WithdrawIntervalSeconds int    `json:",default=10"`
}

// IsEnabled 回傳 Tron 功能是否已設定（熱錢包地址與私鑰皆必填）
func (c TronConf) IsEnabled() bool {
	return c.HotWalletAddress != "" && c.HotWalletPrivateKey != ""
}

// ECPayConf ECPay 金流設定（欄位、環境變數、default 值與 legacy internal/config.ECPayConf 完全一致，
// HashKey/HashIV 透過環境變數注入，不得 commit）。全部欄位皆為 optional，不納入 Validate() 必填檢查。
type ECPayConf struct {
	MerchantID    string `json:",optional"`
	HashKey       string `json:",optional,env=ECPAY_HASH_KEY"`
	HashIV        string `json:",optional,env=ECPAY_HASH_IV"`
	BaseURL       string `json:",default=https://payment-stage.ecpay.com.tw"`
	ReturnURL     string `json:",optional,env=ECPAY_RETURN_URL"`
	ClientBackURL string `json:",optional,env=ECPAY_CLIENT_BACK_URL"`
}

// IsEnabled 回傳 ECPay 功能是否已設定（MerchantID、HashKey、HashIV 皆必填）
func (c ECPayConf) IsEnabled() bool {
	return c.MerchantID != "" && c.HashKey != "" && c.HashIV != ""
}

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
	// Tron / ECPay 皆為 optional 區塊：未設定時 IsEnabled() 為 false，功能關閉但服務仍可啟動，
	// 不加入 Validate() 的必填檢查（本機開發不需要 ECPay/Tron 憑證也要能啟動）。
	Tron  TronConf
	ECPay ECPayConf
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
