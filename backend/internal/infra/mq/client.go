package mq

import (
	"fmt"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/config"
)

type Client struct {
	nc           *nats.Conn
	js           nats.JetStreamContext
	consumerName string
}

func New(c config.NatsConf) *Client {
	if c.URL == "" {
		return nil
	}
	var opts []nats.Option
	if c.CredsPath != "" {
		opts = append(opts, nats.UserCredentials(c.CredsPath))
	}
	if c.User != "" && c.Password != "" {
		opts = append(opts, nats.UserInfo(c.User, c.Password))
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
	nc, err := nats.Connect(c.URL, opts...)
	if err != nil {
		logx.Errorf("nats connect %s error: %v", c.URL, err)
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		logx.Errorf("nats jetstream error: %v", err)
		return nil
	}
	for _, fn := range []func(nats.JetStreamContext) error{
		ensureOrdersStream,
		ensureLedgerStream,
		ensureNotifyStream,
	} {
		if err := fn(js); err != nil {
			logx.Errorf("nats ensure stream error: %v", err)
		}
	}
	logx.Infof("nats jetstream connected: %s", c.URL)
	return &Client{nc: nc, js: js, consumerName: c.ConsumerName}
}

func (c *Client) Ping() error {
	if !c.nc.IsConnected() {
		return fmt.Errorf("nats not connected")
	}
	return nil
}

func (c *Client) Close() {
	c.nc.Drain()
}

// ensureOrdersStream 建立 P2P_ORDERS stream（WorkQueue，7 天保留）。
func ensureOrdersStream(js nats.JetStreamContext) error {
	const streamName = "P2P_ORDERS"
	if _, err := js.StreamInfo(streamName); err == nil {
		return nil
	}
	_, err := js.AddStream(&nats.StreamConfig{
		Name: streamName,
		Subjects: []string{
			"order.*",
			"order.payment.*",
		},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
		MaxAge:    7 * 24 * time.Hour,
		Discard:   nats.DiscardOld,
	})
	return err
}

// ensureLedgerStream 建立 P2P_LEDGER stream（WorkQueue，30 天保留）。
func ensureLedgerStream(js nats.JetStreamContext) error {
	const streamName = "P2P_LEDGER"
	if _, err := js.StreamInfo(streamName); err == nil {
		return nil
	}
	_, err := js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{"ledger.*"},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
		MaxAge:    30 * 24 * time.Hour,
		Discard:   nats.DiscardOld,
	})
	return err
}

// ensureNotifyStream 建立 P2P_NOTIFY stream（WorkQueue，3 天保留）。
// 用於 notify.admin.*（後台 WS 廣播）、notify.buyer.*、notify.seller.*（Push Notification）。
func ensureNotifyStream(js nats.JetStreamContext) error {
	const streamName = "P2P_NOTIFY"
	if _, err := js.StreamInfo(streamName); err == nil {
		return nil
	}
	_, err := js.AddStream(&nats.StreamConfig{
		Name: streamName,
		Subjects: []string{
			"notify.admin.*",
			"notify.buyer.*",
			"notify.seller.*",
		},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
		MaxAge:    3 * 24 * time.Hour,
		Discard:   nats.DiscardOld,
	})
	return err
}
