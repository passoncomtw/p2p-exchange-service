package wsnotification

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
	mu           sync.RWMutex
	userConns    map[int64][]*Conn // app 連線：userID → conns
	backendConns []*Conn           // backend 連線：全體廣播
}

func NewHub() *Hub {
	return &Hub{
		userConns:    make(map[int64][]*Conn),
		backendConns: make([]*Conn, 0),
	}
}

func (h *Hub) Serve(c *Conn) {
	h.register(c)
	go c.writePump()
	c.readPump(h)
}

func (h *Hub) register(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c.Platform == "backend" {
		h.backendConns = append(h.backendConns, c)
	} else {
		h.userConns[c.UserID] = append(h.userConns[c.UserID], c)
	}
}

func (h *Hub) unregister(c *Conn) {
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

func (c *Conn) readPump(h *Hub) {
	defer func() {
		h.unregister(c)
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
				logx.Errorf("ws: write failed for user %d: %v", c.UserID, err)
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
