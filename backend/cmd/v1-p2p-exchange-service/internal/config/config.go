package config

import (
	"github.com/zeromicro/go-zero/core/conf"
)

type Config struct {
	Name string
	Host string
	Port int
	Mode string
}

func NewConfig() *Config {
	var config Config
	if err := conf.Load("etc/config.yaml", &config); err != nil {
		panic(err)
	}
	return &config
}
