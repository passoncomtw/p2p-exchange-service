// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type AuthConfig struct {
	Secret   string
	ExpireIn int64
}

type Config struct {
	rest.RestConf
	Auth AuthConfig
}
