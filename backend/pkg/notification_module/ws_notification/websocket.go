package wsnotification

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,                                       // 讀取緩衝區大小 1024 bytes
	WriteBufferSize: 1024,                                       // 寫入緩衝區大小 1024 bytes
	CheckOrigin:     func(r *http.Request) bool { return true }, // 檢查原始來源
}

type Websocket struct {
	hub *Hub
}

func NewWebsocket(hub *Hub) *Websocket {
	return &Websocket{hub: hub}
}

func (ws *Websocket) ServeWS(w http.ResponseWriter, r *http.Request, userID int64, platform string) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("ws: upgrade failed: %v", err)
		return
	}

	conn := &Conn{
		UserID:   userID,
		Platform: platform,
		conn:     wsConn,
		send:     make(chan []byte, 256),
		hub:      ws.hub,
	}

	ws.hub.Serve(conn)
}
