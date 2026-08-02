package config

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
}

func NewConfig() *Config {
	var config Config
	if err := conf.Load("etc/config.yaml", &config); err != nil {
		panic(err)
	}
	return &config
}
