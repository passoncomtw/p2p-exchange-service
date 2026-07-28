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
	if err := ensureStream(js, c); err != nil {
		logx.Errorf("nats ensure stream error: %v", err)
	}
	if err := ensureNotifyStream(js); err != nil {
		logx.Errorf("nats ensure P2P_NOTIFY stream error: %v", err)
	}
	logx.Infof("nats jetstream connected: %s stream=%s", c.URL, c.StreamName)
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

func ensureStream(js nats.JetStreamContext, c config.NatsConf) error {
	_, err := js.StreamInfo(c.StreamName)
	if err == nil {
		return nil
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name: c.StreamName,
		Subjects: []string{
			"order.timeout.check",
			"order.status.changed",
			"wallet.balance.changed",
			"notification.push",
		},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
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
		MaxAge:    72 * time.Hour,
	})
	return err
}
