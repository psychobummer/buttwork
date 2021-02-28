package device

import (
	"tinygo.org/x/bluetooth"
)

// BLEDevice is a bluetooth implementation of the `Device` interface. This is
// effectively a wrapper around a `*bluetooth.Device`
type BLEDevice struct {
	device         *bluetooth.Device
	tx             bluetooth.DeviceCharacteristic
	rx             bluetooth.DeviceCharacteristic
	implementation Implementation
}

// Vibrate will cause the device to vibrate at the specified level.
func (b BLEDevice) Vibrate(level uint8) (int, error) {
	return b.tx.WriteWithoutResponse(b.implementation.VibrateCommand(level))
}

// Disconnect disconnects the device. In this case it shuts down the underlying bluetooth
// connection
func (b BLEDevice) Disconnect() error {
	return b.device.Disconnect()
}
