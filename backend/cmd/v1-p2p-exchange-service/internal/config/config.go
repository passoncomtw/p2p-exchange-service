package config

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "internal/constants/config.yaml", "the config file")

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
	conf.MustLoad(*configFile, &config)
	return &config
}
