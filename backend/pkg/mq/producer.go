package mq

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// Producer 實作 Publisher interface，封裝 JetStream 發布邏輯。
type Producer struct {
	client *Client
}

func NewProducer(c *Client) *Producer {
	return &Producer{client: c}
}

func (p *Producer) Publish(ctx context.Context, subject string, data []byte) error {
	if p.client == nil || p.client.js == nil {
		return nil
	}
	_, err := p.client.js.Publish(subject, data)
	return err
}

func (p *Producer) PublishAsync(subject string, data []byte) {
	if p.client == nil || p.client.js == nil {
		return
	}
	go func() {
		if _, err := p.client.js.Publish(subject, data); err != nil {
			logx.Errorf("mq publish %s error: %v", subject, err)
		}
	}()
}
