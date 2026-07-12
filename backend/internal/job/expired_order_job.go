package job

import (
	nats "github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/logx"
)

// StartExpiredOrderConsumer 訂閱 order.timeout.check，處理超時訂單。
// 實際掃描與取消邏輯待 PEP-19 實作；此處為 stub，收到後直接 Ack。
func StartExpiredOrderConsumer(js nats.JetStreamContext, consumerName string) {
	if js == nil {
		return
	}
	_, err := js.Subscribe("order.timeout.check", func(msg *nats.Msg) {
		logx.Info("order.timeout.check received (PEP-19 pending)")
		msg.Ack()
	}, nats.Durable(consumerName), nats.ManualAck())
	if err != nil {
		logx.Errorf("subscribe order.timeout.check error: %v", err)
	}
}
