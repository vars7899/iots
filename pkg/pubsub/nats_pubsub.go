package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type NatsPubSub struct {
	conn   *nats.Conn
	subs   map[string]*nats.Subscription
	mu     sync.RWMutex
	logger *zap.Logger
}

func NewNatsPubSub(url string, baseLogger *zap.Logger) (PubSubPublisher, error) {
	if url == "" {
		panic("NatsPubSub: missing dependency 'url'")
	}
	if baseLogger == nil {
		panic("NatsPubSub: missing dependency 'baseLogger'")
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsPubSub{
		conn:   conn,
		subs:   make(map[string]*nats.Subscription),
		logger: logger.Named(baseLogger, "NatsPubSub"),
	}, nil
}

func (n *NatsPubSub) Publish(ctx context.Context, topic string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return n.conn.Publish(topic, data)
}

func (n *NatsPubSub) Subscribe(ctx context.Context, topic string) (<-chan []byte, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, ok := n.subs[topic]; ok {
		return nil, errors.New("already subscribed to topic: " + topic)
	}

	ch := make(chan []byte)

	sub, err := n.conn.Subscribe(topic, func(msg *nats.Msg) {
		select {
		case ch <- msg.Data:
		case <-ctx.Done():
		}
	})
	if err != nil {
		return nil, err
	}

	n.subs[topic] = sub
	return ch, nil
}

func (n *NatsPubSub) Unsubscribe(ctx context.Context, topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	sub, ok := n.subs[topic]
	if !ok {
		return errors.New("not subscribed to topic: " + topic)
	}
	err := sub.Unsubscribe()
	delete(n.subs, topic)
	return err
}

func (n *NatsPubSub) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	for topic, sub := range n.subs {
		sub.Unsubscribe()
		delete(n.subs, topic)
	}
	n.conn.Close()
	return nil
}
