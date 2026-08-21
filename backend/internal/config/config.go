// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"fmt"

	"github.com/zeromicro/go-zero/rest"
)

// AppPlatformConfig App 端 JWT 設定（AccessSecret 敏感，透過環境變數注入）
type AppPlatformConfig struct {
	AccessSecret string `json:",env=APP_JWT_ACCESS_SECRET"`
	AccessExpire int64
}

// BackendPlatformConfig Backend 端 JWT 設定（AccessSecret 敏感，透過環境變數注入）
type BackendPlatformConfig struct {
	AccessSecret string `json:",env=BACKEND_JWT_ACCESS_SECRET"`
	AccessExpire int64
}

// DatabaseConf 資料庫連線設定（DSN 含帳密，透過環境變數注入，不得 commit）
type DatabaseConf struct {
	DSN string `json:",env=DATABASE_DSN"`
}

type RedisConf struct {
	// Mode: single | sentinel | cluster (default: single)
	Mode       string `json:",default=single"`
	Addr       string `json:",env=REDIS_ADDR"`
	Password   string `json:",optional,env=REDIS_PASSWORD"`
	MasterName string `json:",optional"` // sentinel mode only
	PoolSize   int    `json:",default=10"`
}

type NatsConf struct {
	URL          string `json:",env=NATS_URL"`
	CredsPath    string `json:",optional,env=NATS_CREDS_PATH"`
	User         string `json:",optional,env=NATS_USER"`
	Password     string `json:",optional,env=NATS_PASSWORD"`
	StreamName   string
	ConsumerName string
}

// TronConf 平台 Tron 熱錢包設定（Nile Testnet，mainnet 前替換 HotWalletPrivateKey 為 KMS）
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

// ECPayConf ECPay 金流設定（HashKey/HashIV 透過環境變數注入，不得 commit）
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
	App      AppPlatformConfig
	Backend  BackendPlatformConfig
	Database DatabaseConf
	Redis    RedisConf
	Nats     NatsConf
	Tron     TronConf
	ECPay    ECPayConf
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
