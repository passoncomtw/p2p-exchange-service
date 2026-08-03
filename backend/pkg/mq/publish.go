package mq

import "github.com/zeromicro/go-zero/core/logx"

// PublishAsync 在 goroutine 中非同步發布，失敗只記 log（best-effort）。
func (c *Client) PublishAsync(subject string, data []byte) {
	if c == nil || c.js == nil {
		return
	}
	go func() {
		if _, err := c.js.Publish(subject, data); err != nil {
			logx.Errorf("mq publish %s error: %v", subject, err)
		}
	}()
}
