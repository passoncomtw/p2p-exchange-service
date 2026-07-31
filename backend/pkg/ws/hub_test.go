package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// stubConn wraps a channel to record sent bytes, simulating a *websocket.Conn send path.
type stubConn struct {
	sent chan []byte
}

func newStubConn(buf int) *Conn {
	hub := NewHub()
	sc := &Conn{
		UserID:   0,
		Platform: "app",
		conn:     nil, // not needed for hub unit tests
		send:     make(chan []byte, buf),
		hub:      hub,
	}
	return sc
}

func makeConn(hub *Hub, userID int64, platform string) *Conn {
	return &Conn{
		UserID:   userID,
		Platform: platform,
		conn:     nil,
		send:     make(chan []byte, 64),
		hub:      hub,
	}
}

func TestHubRegisterAndConnect(t *testing.T) {
	hub := NewHub()
	c := makeConn(hub, 42, "app")
	hub.Register(c)

	hub.mu.RLock()
	conns := hub.userConns[42]
	hub.mu.RUnlock()

	if len(conns) != 1 {
		t.Fatalf("expected 1 conn for userID 42, got %d", len(conns))
	}
}

func TestHubMultiDeviceConnect(t *testing.T) {
	hub := NewHub()
	c1 := makeConn(hub, 7, "app")
	c2 := makeConn(hub, 7, "app")
	hub.Register(c1)
	hub.Register(c2)

	hub.mu.RLock()
	conns := hub.userConns[7]
	hub.mu.RUnlock()

	if len(conns) != 2 {
		t.Fatalf("expected 2 conns for userID 7 (multi-device), got %d", len(conns))
	}
}

func TestHubSendToUser(t *testing.T) {
	hub := NewHub()
	c := makeConn(hub, 1, "app")
	hub.Register(c)

	msg := []byte(`{"type":"order.status.changed"}`)
	hub.SendToUser(1, msg)

	select {
	case received := <-c.send:
		if string(received) != string(msg) {
			t.Fatalf("expected %q, got %q", msg, received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for message")
	}
}

func TestHubBroadcastToBackend(t *testing.T) {
	hub := NewHub()
	b1 := makeConn(hub, 100, "backend")
	b2 := makeConn(hub, 101, "backend")
	hub.Register(b1)
	hub.Register(b2)

	msg := []byte(`{"type":"order.status.changed"}`)
	hub.BroadcastToBackend(msg)

	for i, c := range []*Conn{b1, b2} {
		select {
		case received := <-c.send:
			if string(received) != string(msg) {
				t.Fatalf("backend conn %d: expected %q, got %q", i, msg, received)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timed out waiting for broadcast on backend conn %d", i)
		}
	}
}

func TestHubDisconnectReleasesResources(t *testing.T) {
	hub := NewHub()
	c := makeConn(hub, 9, "app")
	hub.Register(c)
	hub.Unregister(c)

	hub.mu.RLock()
	_, exists := hub.userConns[9]
	hub.mu.RUnlock()

	if exists {
		t.Fatal("expected userID 9 to be removed from hub after disconnect")
	}

	// send channel should be closed
	_, open := <-c.send
	if open {
		t.Fatal("expected send channel to be closed after Unregister")
	}
}

func TestHubDisconnectOneOfMultiDevice(t *testing.T) {
	hub := NewHub()
	c1 := makeConn(hub, 5, "app")
	c2 := makeConn(hub, 5, "app")
	hub.Register(c1)
	hub.Register(c2)

	hub.Unregister(c1)

	hub.mu.RLock()
	conns := hub.userConns[5]
	hub.mu.RUnlock()

	if len(conns) != 1 {
		t.Fatalf("expected 1 remaining conn for userID 5 after one disconnect, got %d", len(conns))
	}
	if conns[0] != c2 {
		t.Fatal("expected remaining conn to be c2")
	}
}

func TestNewMessage(t *testing.T) {
	payload := OrderStatusChangedData{
		OrderID:  123,
		Status:   "completed",
		BuyerID:  1,
		SellerID: 2,
	}
	raw, err := NewMessage(EventOrderStatusChanged, payload)
	if err != nil {
		t.Fatalf("NewMessage error: %v", err)
	}
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if msg.Type != EventOrderStatusChanged {
		t.Fatalf("expected type %q, got %q", EventOrderStatusChanged, msg.Type)
	}
}

// Ensure websocket import is available (compile-time).
var _ = websocket.CloseMessage
