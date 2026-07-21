package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 512
)

// Conn 包裝單一 WebSocket 連線。
type Conn struct {
	UserID   int64
	Platform string // "app" | "backend"
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
}

// Hub 管理所有 WebSocket 連線，提供依用戶 ID 發送與全體廣播功能。
type Hub struct {
	mu             sync.RWMutex
	userConns      map[int64][]*Conn // app 連線：userID → conns
	backendConns   []*Conn          // backend 連線：全體廣播
}

func NewHub() *Hub {
	return &Hub{
		userConns:    make(map[int64][]*Conn),
		backendConns: make([]*Conn, 0),
	}
}

func (h *Hub) Register(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c.Platform == "backend" {
		h.backendConns = append(h.backendConns, c)
	} else {
		h.userConns[c.UserID] = append(h.userConns[c.UserID], c)
	}
}

func (h *Hub) Unregister(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c.Platform == "backend" {
		h.backendConns = removeConn(h.backendConns, c)
	} else {
		conns := removeConn(h.userConns[c.UserID], c)
		if len(conns) == 0 {
			delete(h.userConns, c.UserID)
		} else {
			h.userConns[c.UserID] = conns
		}
	}
	close(c.send)
}

// SendToUser 非同步推送訊息給指定用戶的所有連線（App）。
func (h *Hub) SendToUser(userID int64, msg []byte) {
	h.mu.RLock()
	conns := h.userConns[userID]
	h.mu.RUnlock()
	for _, c := range conns {
		select {
		case c.send <- msg:
		default:
			logx.Errorf("ws: send buffer full for user %d, dropping message", userID)
		}
	}
}

// BroadcastToBackend 非同步推送訊息給所有後台連線。
func (h *Hub) BroadcastToBackend(msg []byte) {
	h.mu.RLock()
	conns := make([]*Conn, len(h.backendConns))
	copy(conns, h.backendConns)
	h.mu.RUnlock()
	for _, c := range conns {
		select {
		case c.send <- msg:
		default:
			logx.Errorf("ws: send buffer full for backend conn, dropping message")
		}
	}
}

// Serve 啟動 readPump 和 writePump goroutine，並在完成時從 Hub 移除。
func (h *Hub) Serve(c *Conn) {
	h.Register(c)
	go c.writePump()
	c.readPump(h)
}

// NewConn 建立新的 Conn 並關聯 Hub。
func NewConn(userID int64, platform string, wsConn *websocket.Conn, hub *Hub) *Conn {
	return &Conn{
		UserID:   userID,
		Platform: platform,
		conn:     wsConn,
		send:     make(chan []byte, 256),
		hub:      hub,
	}
}

// readPump 持續讀取客戶端訊息（pong）直到連線關閉。
func (c *Conn) readPump(h *Hub) {
	defer func() {
		h.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump 從 send channel 取訊息寫入 WebSocket，並定期發送 ping。
func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func removeConn(conns []*Conn, target *Conn) []*Conn {
	result := conns[:0]
	for _, c := range conns {
		if c != target {
			result = append(result, c)
		}
	}
	return result
}
