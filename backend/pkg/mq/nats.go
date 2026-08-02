package mq

import (
	"p2p-exchange/internal/config"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
)

type Client struct {
	nc           *nats.Conn
	js           nats.JetStreamContext
	consumerName string
}

func NewNats(config config.Config) *Client {
	var opts []nats.Option
	switch {
	case config.Nats.CredsPath != "":
		opts = append(opts, nats.UserCredentials(config.Nats.CredsPath))
	case config.Nats.User != "" && config.Nats.Password != "":
		opts = append(opts, nats.UserInfo(config.Nats.User, config.Nats.Password))
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

	nc, err := nats.Connect(config.Nats.URL, opts...)
	if err != nil {
		logx.Errorf("nats connect %s error: %v", config.Nats.URL, err)
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		logx.Errorf("nats jetstream error: %v", err)
		return nil
	}

	logx.Infof("nats jetstream connected: %s", config.Nats.URL)
	return &Client{nc: nc, js: js, consumerName: config.Nats.ConsumerName}
}
