package config

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Database struct {
		DSN string `json:"dsn"`
	} `json:"database"`
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		PoolSize int    `json:"poolSize"`
	} `json:"redis"`
	WebSocket struct {
		Enabled bool   `json:"enabled"`
		Host    string `json:"host"`
		Port    int    `json:"port"`
	} `json:"webSocket"`
}

func NewConfig() *Config {
	var config Config
	if err := conf.Load("cmd/v1-p2p-exchange-service/internal/constants/config.yaml", &config); err != nil {
		panic(err)
	}
	return &config
}
