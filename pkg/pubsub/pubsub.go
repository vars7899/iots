package pubsub

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type PubSubPublisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
	Unsubscribe(ctx context.Context, topic string) error
	Close() error
}

const (
	// Events from devices
	NatsTopicDeviceEvents = "device.events"
	// Topic for incoming commands from the WebSocket handler or other sources
	NatsTopicCommandsInbound = "commands.inbound"
	// Topic prefix for outgoing responses/events to specific devices
	// Responses/events will be published to "commands.outbound.<deviceID>"
	NatsTopicCommandsOutboundPrefix = "commands.outbound."
	// Topic prefix for general device state change events
	// Events will be published to "device.events.<deviceID>"
	NatsTopicDeviceEventsPrefix = "device.events."
	// System events (errors, monitoring, etc.)
	NatsTopicSystemEvents = "system.events"
)

func NatsTopicCommandsOutboundPrefixf(deviceID uuid.UUID) string {
	return fmt.Sprintf("%s%s", NatsTopicCommandsOutboundPrefix, deviceID.String())
}

func NatsTopicDeviceEventsPrefixf(deviceID uuid.UUID) string {
	return fmt.Sprintf("%s%s", NatsTopicDeviceEventsPrefix, deviceID.String())
}
