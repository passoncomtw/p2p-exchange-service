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

type Config struct {
	rest.RestConf
	App      PlatformConfig
	Backend  PlatformConfig
	Database DatabaseConf
	Redis    RedisConf
	Nats     NatsConf
}
