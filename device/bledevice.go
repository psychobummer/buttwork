package device

import (
	"fmt"

	"tinygo.org/x/bluetooth"
)

// BLEDevice is a bluetooth implementation of the `Device` interface. This is
// effectively a wrapper around a `*bluetooth.Device`
type BLEDevice struct {
	device *bluetooth.Device
	tx     bluetooth.DeviceCharacteristic
	rx     bluetooth.DeviceCharacteristic
}

// Vibrate will cause the device to vibrate at the specified level.
func (b BLEDevice) Vibrate(level uint8) (int, error) {
	strCommand := fmt.Sprintf("Vibrate:%d;\n", level)
	command := []byte(strCommand)
	return b.tx.WriteWithoutResponse(command)

}

// Disconnect disconnects the device. In this case it shuts down the underlying bluetooth
// connection
func (b BLEDevice) Disconnect() error {
	return b.device.Disconnect()
}
