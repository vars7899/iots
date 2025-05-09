package websocket

import "fmt"

type EventType string

const (
	EventTelemetry EventType = "telemetry"
	EventAlert     EventType = "alert"
)

type WebSocketEvent struct {
	Type    EventType
	Payload []byte
}

func TriggerEvent(session *DeviceSession, event WebSocketEvent) {
	msg := fmt.Sprintf("%s: %s", event.Type, event.Payload)
	session.Channel <- []byte(msg)
}
