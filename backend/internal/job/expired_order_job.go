package job

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/infra/mq"
)

// StartExpiredOrderConsumer 訂閱 order.timeout.check，處理超時訂單。
// 實際掃描與取消邏輯待 PEP-19 實作；此處為 stub，收到後直接 Ack。
func StartExpiredOrderConsumer(mqClient *mq.Client) {
	if mqClient == nil {
		return
	}
	if err := mqClient.Subscribe("order.timeout.check", func(_ context.Context, _ []byte) error {
		logx.Info("order.timeout.check received (PEP-19 pending)")
		return nil
	}); err != nil {
		logx.Errorf("subscribe order.timeout.check error: %v", err)
	}
}
