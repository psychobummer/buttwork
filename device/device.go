package device

// Device defines a contract all devices must satisfy.
type Device interface {
	Vibrate(level uint8) (int, error)
	Disconnect() error
}
