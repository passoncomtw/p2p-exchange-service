package mq

import (
	"context"
	"testing"

	nats "github.com/nats-io/nats.go"
)

// stubJetStream wraps a recorded publish call to test PublishAsync indirectly.
type recordedPub struct {
	subject string
	data    []byte
}

func TestPublishAsync_CallsJetStreamPublish(t *testing.T) {
	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("no local NATS server: " + err.Error())
	}
	defer ns.Close()

	js, err := ns.JetStream()
	if err != nil {
		t.Fatal(err)
	}

	const stream = "TEST_PUBLISHER"
	_ = js.DeleteStream(stream)
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      stream,
		Subjects:  []string{"test.pub.*"},
		Storage:   nats.MemoryStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil {
		t.Fatal("AddStream:", err)
	}
	defer js.DeleteStream(stream)

	c := &Client{nc: ns, js: js, consumerName: "test"}
	msg := []byte(`{"test":true}`)

	if err := c.Publish(context.Background(), "test.pub.foo", msg); err != nil {
		t.Fatalf("Publish error: %v", err)
	}

	info, err := js.StreamInfo(stream)
	if err != nil {
		t.Fatal(err)
	}
	if info.State.Msgs != 1 {
		t.Fatalf("expected 1 message in stream, got %d", info.State.Msgs)
	}
}

func TestPublishAsync_BestEffort(t *testing.T) {
	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("no local NATS server: " + err.Error())
	}
	defer ns.Close()

	js, err := ns.JetStream()
	if err != nil {
		t.Fatal(err)
	}

	const stream = "TEST_ASYNC_PUB"
	_ = js.DeleteStream(stream)
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      stream,
		Subjects:  []string{"test.async.*"},
		Storage:   nats.MemoryStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil {
		t.Fatal("AddStream:", err)
	}
	defer js.DeleteStream(stream)

	c := &Client{nc: ns, js: js, consumerName: "test"}
	c.PublishAsync("test.async.bar", []byte(`{"async":true}`))

	// give the goroutine time to publish
	for i := 0; i < 20; i++ {
		info, _ := js.StreamInfo(stream)
		if info != nil && info.State.Msgs == 1 {
			return
		}
	}
	t.Fatal("PublishAsync: message not delivered within timeout")
}

func TestPing_Connected(t *testing.T) {
	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("no local NATS server: " + err.Error())
	}
	defer ns.Close()

	c := &Client{nc: ns}
	if err := c.Ping(); err != nil {
		t.Fatalf("Ping() on connected client returned error: %v", err)
	}
}
