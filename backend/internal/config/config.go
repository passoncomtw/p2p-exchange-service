// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type PlatformConfig struct {
	AccessSecret string
	AccessExpire int64
}

type DatabaseConf struct {
	DSN string
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
	CredsPath    string `json:",optional"`
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

type Config struct {
	rest.RestConf
	App      PlatformConfig
	Backend  PlatformConfig
	Database DatabaseConf
	Redis    RedisConf
	Nats     NatsConf
	Tron     TronConf
}
