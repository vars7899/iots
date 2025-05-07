package websocket

import "fmt"

const (
	TopicDeviceRegistered = "device.registered.%s"
	TopicDeviceMetadata   = "device.metadata.%s"
	TopicDeviceState      = "device.state.%s"
)

func formatTopic(format, key string) string {
	return fmt.Sprintf(format, key)
}
