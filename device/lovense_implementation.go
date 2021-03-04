package device

import "fmt"

// LovenseImplementation provides a concrete implmentation of the Lovense protocol
type LovenseImplementation struct{}

// Init is a NOOP for lovense devices
func (l LovenseImplementation) Init() error {
	return nil
}

// Name returns the implementation name
func (l LovenseImplementation) Name() string {
	return "lovense"
}

// VibrateCommand will vibrate a lovense device
func (l LovenseImplementation) VibrateCommand(level uint8) []byte {
	deviceMax := uint8(20)
	normalized := normalize(level, deviceMax)
	command := fmt.Sprintf("Vibrate:%d;\n", normalized)
	return []byte(command)
}
