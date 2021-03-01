package device

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"
)

// Discovery provides a means of discovering devices.
type Discovery interface {
	Connect(Identifier) (Device, error)
	ScanOnce(duration time.Duration) (Identifiers, error)
	ScanBackground() (<-chan Identifier, <-chan error)
}

// BLEDiscovery is a BLE implementation of the Discovery interface.
// Allows for the discovery of, and connection to, bluetooth devices.
type BLEDiscovery struct {
	adapter *bluetooth.Adapter
}

// NewBLEDiscovery will initialize bluetooth and return
// something you can use to perform discovery.
func NewBLEDiscovery() (Discovery, error) {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return nil, err
	}
	return BLEDiscovery{adapter: adapter}, nil
}

// ScanBackground forks a gofunc and returns a constant stream of Identifiers for
// any devices it has located, and any errors. It returns channels you can use to
// collect this data.
func (d BLEDiscovery) ScanBackground() (<-chan Identifier, <-chan error) {
	idents := make(chan Identifier)
	errors := make(chan error)
	go func(idents chan Identifier, errors chan error) {
		err := d.adapter.Scan(func(adapter *bluetooth.Adapter, found bluetooth.ScanResult) {
			ident := Identifier{
				Address: found.Address.String(),
				Name:    found.LocalName(),
			}
			idents <- ident
		})
		if err != nil {
			errors <- err
		}
	}(idents, errors)
	return idents, errors
}

// ScanOnce will scan for the specified `time.Duration` and return a list of any `Identifier`s
// it saw.
func (d BLEDiscovery) ScanOnce(duration time.Duration) (Identifiers, error) {
	deduplication := make(map[string]Identifier)
	idents, errors := d.ScanBackground()
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case ident := <-idents:
				if _, found := deduplication[ident.Address]; found {
					continue
				}
				deduplication[ident.Address] = ident
			}
		}
	}()

	select {
	case <-time.After(duration):
		d.StopScan()
		done <- true
	case err := <-errors:
		d.StopScan()
		done <- true
		return nil, err
	}

	results := []Identifier{}
	for _, v := range deduplication {
		log.Debug().Msgf("found BLE device: %v", v)
		results = append(results, v)
	}
	return results, nil
}

// StopScan stops the bluetooth scan process
func (d BLEDiscovery) StopScan() {
	d.adapter.StopScan()
}

// Connect will connect to the `Device` for the device identified by `Identifier`,
// It will connect to the device, discover characteristics, and make determinations about
// tx/rx channels.
func (d BLEDiscovery) Connect(identifier Identifier) (Device, error) {
	log.Debug().Msgf("attempting to connect to identifier: %v", identifier)
	uuid, err := bluetooth.ParseUUID(identifier.Address)
	if err != nil {
		return nil, err
	}

	address := bluetooth.Address{
		UUID: uuid,
	}
	device, err := d.adapter.Connect(address, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, err
	}

	// nil argument gets all services
	services, err := device.DiscoverServices(nil)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("Device at %s has no registered services", identifier.Address)
	}
	service := services[0]

	// nil argument gets all characteristics
	characteristics, err := service.DiscoverCharacteristics(nil)
	if err != nil {
		return nil, err
	}

	// TODO identify these better. right now ordering is sufficient.
	retDevice := BLEDevice{
		device:         device,
		tx:             characteristics[0],
		rx:             characteristics[1],
		implementation: NewImplementation(identifier.Name),
	}

	return retDevice, nil

}
