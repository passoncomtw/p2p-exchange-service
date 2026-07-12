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
	Addr     string // comma-separated cluster node addresses, or single addr for standalone
	Password string
	PoolSize int
}

type NatsConf struct {
	URL          string
	CredsPath    string // optional: path to .creds file (Synadia Cloud prod)
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
