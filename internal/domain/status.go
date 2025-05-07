package domain

type Status string

var (
	Online        Status = "online"
	Offline       Status = "offline"
	Faulty        Status = "faulty"
	Idle          Status = "idle"
	Disconnected  Status = "disconnected"
	InMaintenance Status = "in_maintenance"
	Pending       Status = "pending"
	Suspended     Status = "suspended"
	Error         Status = "error"
	Activated     Status = "activated"
	Reboot        Status = "reboot"
)

func IsValidStatus(inputStr string) bool {
	switch Status(inputStr) {
	case Online, Offline, Faulty, Idle, Disconnected, InMaintenance, Pending, Suspended, Error, Activated:
		return true
	default:
		return false
	}
}
