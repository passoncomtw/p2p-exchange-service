package mq

import (
	"context"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/fx"
)

type Config struct {
	URL          string
	CredsPath    string
	User         string
	Password     string
	StreamName   string
	ConsumerName string
}

type Client struct {
	nc           *nats.Conn
	js           nats.JetStreamContext
	consumerName string
}

func NewNats(cfg Config, lc fx.Lifecycle) *Client {
	if cfg.URL == "" {
		return nil
	}
	var opts []nats.Option
	switch {
	case cfg.CredsPath != "":
		opts = append(opts, nats.UserCredentials(cfg.CredsPath))
	case cfg.User != "" && cfg.Password != "":
		opts = append(opts, nats.UserInfo(cfg.User, cfg.Password))
	}
	opts = append(opts,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logx.Errorf("nats disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logx.Infof("nats reconnected: %s", nc.ConnectedUrl())
		}),
	)

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		logx.Errorf("nats connect %s error: %v", cfg.URL, err)
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		logx.Errorf("nats jetstream error: %v", err)
		nc.Close()
		return nil
	}

	logx.Infof("nats jetstream connected: %s", cfg.URL)
	client := &Client{nc: nc, js: js, consumerName: cfg.ConsumerName}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			logx.Info("nats: draining connection")
			return nc.Drain()
		},
	})

	return client
}
