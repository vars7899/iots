package domain

// Device Connection
type ConnectionType string

var (
	BluetoothConnection ConnectionType = "bluetooth"
	EthernetConnection  ConnectionType = "ethernet"
	WifiConnection      ConnectionType = "wifi"
	CellularConnection  ConnectionType = "cellular"
	ZigbeeConnection    ConnectionType = "zigbee"
	ZWaveConnection     ConnectionType = "zwave"
	LoRaConnection      ConnectionType = "lora"
	SerialConnection    ConnectionType = "serial"
	USBConnection       ConnectionType = "usb"
	SatelliteConnection ConnectionType = "satellite"
	NFCConnection       ConnectionType = "nfc"
	ThreadConnection    ConnectionType = "thread"
	InfraredConnection  ConnectionType = "infrared"
	UnknownConnection   ConnectionType = "unknown"
)

func IsValidConnectionType(inputStr string) bool {
	switch ConnectionType(inputStr) {
	case BluetoothConnection, EthernetConnection, WifiConnection, CellularConnection, ZigbeeConnection, ZWaveConnection, LoRaConnection, SerialConnection, USBConnection, SatelliteConnection, NFCConnection, ThreadConnection, InfraredConnection, UnknownConnection:
		return true
	default:
		return false
	}
}
