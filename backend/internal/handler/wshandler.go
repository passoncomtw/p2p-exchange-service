package handler

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	pkgws "p2p-exchange/pkg/ws"
	"p2p-exchange/internal/svc"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:    func(r *http.Request) bool { return true },
}

// AppWSHandler 處理 App 端 WebSocket 連線（/ws/app?token=<app_jwt>）。
func AppWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseJWTUID(r, svcCtx.Config.App.AccessSecret)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("ws app upgrade error: %v", err)
			return
		}
		conn := pkgws.NewConn(uid, "app", wsConn, svcCtx.Hub)
		go svcCtx.Hub.Serve(conn)
	}
}

// BackendWSHandler 處理後台 WebSocket 連線（/ws/backend?token=<backend_jwt>）。
func BackendWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := parseJWTUID(r, svcCtx.Config.Backend.AccessSecret)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("ws backend upgrade error: %v", err)
			return
		}
		conn := pkgws.NewConn(uid, "backend", wsConn, svcCtx.Hub)
		go svcCtx.Hub.Serve(conn)
	}
}

// parseJWTUID 從 query param ?token= 或 Authorization header 解析 JWT，回傳 uid。
func parseJWTUID(r *http.Request, secret string) (int64, bool) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		auth := r.Header.Get("Authorization")
		tokenStr = strings.TrimPrefix(auth, "Bearer ")
	}
	if tokenStr == "" {
		return 0, false
	}
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
