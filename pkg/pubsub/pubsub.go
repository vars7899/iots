package pubsub

import "context"

type PubSubPublisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
	Unsubscribe(ctx context.Context, topic string) error
	Close() error
}
