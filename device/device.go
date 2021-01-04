package device

import (
	"strings"
	"time"
)

// Identifier identifies a device by address and, optionally, a name.
// e.g: LVS-*
type Identifier struct {
	Address string
	Name    string
}

// Identifiers is container for multiple `Identifier` structs.
type Identifiers []Identifier

// FilterPrefix will filter a set of Identifiers by name
// For example, if you only wanted devices matching "LVS-*", wantPrefix would be "LVS"
func (i Identifiers) FilterPrefix(wantPrefix string) Identifiers {
	var result []Identifier
	for _, device := range i {
		parts := strings.Split(device.Name, "-")
		if len(parts) > 0 && parts[0] == wantPrefix {
			result = append(result, device)
		}
	}
	return result
}

// Device defines a contract all devices must satisfy.
type Device interface {
	Vibrate(level uint8) (int, error)
	Disconnect() error
}

// Discovery provides a means of discovering devices.
type Discovery interface {
	Connect(Identifier) (Device, error)
	ScanOnce(duration time.Duration) (Identifiers, error)
	ScanBackground() (<-chan Identifier, <-chan error)
}
