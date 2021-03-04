package device

import "fmt"

// Implementation defines the interface a concrete device implementation
type Implementation interface {
	Init() error                 // some devices may require an initialization procedure
	Name() string                // return implementatin name for logging
	VibrateCommand(uint8) []byte // a command that will vibrate the device
}

// NewImplementation returns a concrete implementation for the type of device
// identified by `name`
func NewImplementation(deviceType string) (Implementation, error) {
	switch deviceType {
	case "lovense":
		return LovenseImplementation{}, nil
	case "wevibe":
		return WevibeImplementation{}, nil
	}
	return nil, fmt.Errorf("No implementation defined for type %s", deviceType)
}

// To vibrate a device, you generally say "vibrate at x", where x is some integer value.
// The maximum value, the 100% intensity value, seems generally arbitrary and varies across devices.
// Let's make some attempt at unifying this.
// The following function will normalize a vibration level on the scale [0,100] to [0,deviceMax]
// e,g: normalize(100, 20) => 20; normalize(50, 20) => 10
func normalize(level uint8, deviceMax uint8) uint8 {
	if level > 100 {
		return deviceMax
	}
	if level <= 0 {
		return 0
	}
	clampMax := float32(100)
	vibMax := float32(deviceMax)
	val := vibMax * (float32(level) / clampMax)
	return uint8(val)
}
