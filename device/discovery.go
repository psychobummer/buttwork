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
	config  Config
}

// NewBLEDiscovery will initialize bluetooth and return
// something you can use to perform discovery.
func NewBLEDiscovery(config Config) (Discovery, error) {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return nil, err
	}
	discovery := BLEDiscovery{
		adapter: adapter,
		config:  config,
	}
	return discovery, nil
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
			if ident.Name == "" {
				log.Debug().Msgf("scan result: %v [ignored]", ident)
				return
			}
			log.Debug().Msgf("scan result: %v", ident)
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
		log.Debug().Msgf("found well-formed BLE device: %v", v)
		results = append(results, v)
	}
	return results, nil
}

// StopScan stops the bluSetooth scan process
func (d BLEDiscovery) StopScan() {
	d.adapter.StopScan()
}

// Connect will connect to the `Device` for the device identified by `Identifier`,
// It will connect to the device, discover characteristics, and make determinations about
// tx/rx channels.
func (d BLEDiscovery) Connect(identifier Identifier) (Device, error) {

	if identifier.Name == "" {
		return nil, fmt.Errorf("Device does not specify a name; cannot determine driver")
	}

	if found, _ := d.config.GetTypeFromIdentifier(identifier); !found {
		return nil, fmt.Errorf("config does not specify a driver type for device identified by %s", identifier.Name)
	}

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

	// From here on out, we have established a connection to the device.
	// If we encounter an error, make sure we disconnect from it.
	service, err := d.getSupportedService(identifier, device)
	if err != nil {
		device.Disconnect()
		return nil, err
	}
	log.Debug().Msgf("discovered supported service: %v", service)

	tx, rx, err := d.getSupportedCharacteristics(identifier, service)
	if err != nil {
		device.Disconnect()
		return nil, err
	}
	log.Debug().Msgf("discovered characteristics (tx: %v, rx: %v)", tx, rx)

	// TODO identify these better. right now ordering is sufficient.
	_, deviceType := d.config.GetTypeFromIdentifier(identifier)
	implementation, err := NewImplementation(deviceType)
	if err != nil {
		device.Disconnect()
		return nil, err
	}
	retDevice := BLEDevice{
		device:         device,
		tx:             tx,
		rx:             rx,
		implementation: implementation,
	}

	// caller is responsible for disconnecting now
	return retDevice, nil
}

// GetSupportedService will enumerate all services expose by the device, and return the first one that
// we know how to support. As is we don't have any devices in our posession that support multiple services.
// TODO: tidy up; do all config-specified UUID validation up front
func (d BLEDiscovery) getSupportedService(identifier Identifier, device *bluetooth.Device) (bluetooth.DeviceService, error) {
	// nil argument gets all services
	discoveredServices, err := device.DiscoverServices(nil)
	if err != nil {
		return bluetooth.DeviceService{}, err
	}
	for _, supportedService := range d.config.GetServicesForIdentifier(identifier) {
		parsedUUID, err := bluetooth.ParseUUID(supportedService.ID)
		if err != nil {
			log.Error().Msgf("configuration for identifier %s contains an invalid UUID %s: %s", identifier, supportedService.ID, err)
			continue
		}
		for _, discoveredService := range discoveredServices {
			if parsedUUID == discoveredService.UUID() {
				// we can support this discovered service
				return discoveredService, nil
			}
		}
	}
	return bluetooth.DeviceService{}, fmt.Errorf("Device %s exposes no supported services", identifier.Name)
}

// GetSupportedCharacteristics will enumerate all characteristics exposed by the given service,
// and return them as a tx/rx touple. The devices we have only expose tx/rx characteristics, but
// who knows if they have a writeable characteristic that, like, wipes the firmware or something,
// so let's try to be safe and only talk to things we know about.
func (d BLEDiscovery) getSupportedCharacteristics(identifier Identifier, service bluetooth.DeviceService) (bluetooth.DeviceCharacteristic, bluetooth.DeviceCharacteristic, error) {

	// nil argument gets all characteristics
	discoveredCharacteristics, err := service.DiscoverCharacteristics(nil)
	if err != nil {
		return bluetooth.DeviceCharacteristic{}, bluetooth.DeviceCharacteristic{}, err
	}

	// look up the characteristic details we have on file for this service
	for _, supportedService := range d.config.GetServicesForIdentifier(identifier) {
		// we found the service we're asking about
		if supportedService.ID == service.String() {
			var tx bluetooth.DeviceCharacteristic
			var rx bluetooth.DeviceCharacteristic
			for _, discovered := range discoveredCharacteristics {
				if supportedService.Tx == discovered.String() {
					tx = discovered
				}
				if supportedService.Rx == discovered.String() {
					rx = discovered
				}
			}
			// we have tx and rx characteristics matching what we've discovered
			if tx.String() != "" && rx.String() != "" {
				return tx, rx, nil
			}
		}
	}
	notFoundErr := fmt.Errorf("Device did not advertise any known characteristics")
	return bluetooth.DeviceCharacteristic{}, bluetooth.DeviceCharacteristic{}, notFoundErr
}
