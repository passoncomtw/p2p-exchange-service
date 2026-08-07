package wsserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/fx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	wsnotification "p2p-exchange/pkg/notification_module/ws_notification"
)

// parseToken 嘗試以 secret 驗證 JWT，回傳 uid。
func parseToken(tokenStr, secret string) (int64, bool) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return 0, false
	}
	return int64(uidFloat), true
}

// extractToken 從 ?token= 或 Authorization header 取出 token 字串。
func extractToken(r *http.Request) string {
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func New(c *config.Config, ws *wsnotification.Websocket, lc fx.Lifecycle) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws/app", func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseToken(extractToken(r), c.App.AccessSecret)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ws.ServeWS(w, r, uid, "app")
	})

	mux.HandleFunc("/ws/backend", func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseToken(extractToken(r), c.Backend.AccessSecret)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ws.ServeWS(w, r, uid, "backend")
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
