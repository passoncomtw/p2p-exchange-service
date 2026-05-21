// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type PlatformConfig struct {
	AccessSecret string
	AccessExpire int64
}

type Config struct {
	rest.RestConf
	App     PlatformConfig
	Backend PlatformConfig
}
