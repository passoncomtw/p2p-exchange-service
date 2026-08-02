package wsserver

import (
	"context"
	"fmt"
	"net/http"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	wsnotification "p2p-exchange/pkg/notification_module/ws_notification"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/fx"
)

func New(c *config.Config, ws *wsnotification.Websocket, lc fx.Lifecycle) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// TODO: 解析 JWT，取得 userID / platform
		// 暫時以 query string 帶入，待 JWT middleware 實作後替換
		userID := int64(0)
		platform := r.URL.Query().Get("platform")
		if platform == "" {
			platform = "app"
		}
		ws.ServeWS(w, r, userID, platform)
	})

	addr := fmt.Sprintf("%s:%d", c.WebSocket.Host, c.WebSocket.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			if !c.WebSocket.Enabled {
				logx.Info("wsserver: disabled, skipping")
				return nil
			}
			go func() {
				logx.Infof("wsserver: listening on %s", addr)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logx.Errorf("wsserver: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logx.Info("wsserver: shutting down")
			return server.Shutdown(ctx)
		},
	})

	return server
}
