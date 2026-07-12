package mq

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

func (c *Client) Publish(_ context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(subject, data)
	return err
}

// PublishAsync 在 goroutine 中非同步發布，失敗只記 log（best-effort）。
func (c *Client) PublishAsync(subject string, data []byte) {
	go func() {
		if _, err := c.js.Publish(subject, data); err != nil {
			logx.Errorf("mq publish %s error: %v", subject, err)
		}
	}()
}
