package ws

import "time"

type SensorTelemetryWsMessage struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
