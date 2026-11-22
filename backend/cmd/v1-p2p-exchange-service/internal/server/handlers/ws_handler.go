package handlers

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	pkgws "p2p-exchange/pkg/ws"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSHandler struct {
	config *config.Config
	hub    *pkgws.Hub
}

func NewWSHandler(c *config.Config, hub *pkgws.Hub) *WSHandler {
	return &WSHandler{config: c, hub: hub}
}

// AppWS 處理 App 端 WebSocket 連線（/ws/app?token=<app_jwt>）。
//
//	@Summary		App WebSocket 連線
//	@Description	以 query 的 token 驗證（App JWT），驗證成功後升級為 WebSocket。
//	@Tags			websocket
//	@Param			token	query		string	true	"App JWT"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Failure		401		{string}	string	"unauthorized"
//	@Router			/ws/app [get]
func (h *WSHandler) AppWS(w http.ResponseWriter, r *http.Request) {
	uid, ok := parseWSJWT(r, h.config.App.AccessSecret)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	wsConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("ws app upgrade: %v", err)
		return
	}
	conn := pkgws.NewConn(uid, "app", wsConn, h.hub)
	go h.hub.Serve(conn)
}

// BackendWS 處理後台 WebSocket 連線（/ws/backend?token=<backend_jwt>）。
//
//	@Summary		後台 WebSocket 連線
//	@Description	以 query 的 token 驗證（Backend JWT），驗證成功後升級為 WebSocket。
//	@Tags			websocket
//	@Param			token	query		string	true	"Backend JWT"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Failure		401		{string}	string	"unauthorized"
//	@Router			/ws/backend [get]
func (h *WSHandler) BackendWS(w http.ResponseWriter, r *http.Request) {
	uid, ok := parseWSJWT(r, h.config.Backend.AccessSecret)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	wsConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("ws backend upgrade: %v", err)
		return
	}
	conn := pkgws.NewConn(uid, "backend", wsConn, h.hub)
	go h.hub.Serve(conn)
}

// parseWSJWT 從 ?token= 或 Authorization: Bearer 解析 JWT，回傳 uid。
func parseWSJWT(r *http.Request, secret string) (int64, bool) {
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
