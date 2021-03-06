package device

import (
	"regexp"

	"github.com/rs/zerolog/log"
)

// Config represents the top-level mapping of device configurations
/*
TODO: make this json. i hate yaml.
drivers:
  -
    type: lovense
    names:
      - LVS-.*
      - LOVE-.*
    services:
      -
        rx: 5a300003-0023-4bd4-bbd5-a6920e4c5653
        id: 5a300001-0023-4bd4-bbd5-a6920e4c5653
        tx: 5a300002-0023-4bd4-bbd5-a6920e4c5653
*/
type Config struct {
	Drivers []Driver `yaml:"drivers"`
}

// A Driver defines how we'll communicate with a discovered device
// `Type` determines which `Implementation` to use; ie: "lovense", "kiiroo-v2"
// `Names` contains a set of regular expression we'll look for in the GAP section of the GATT announcement
// `Services` contains a set of known service UUIDs, and the tx/rx characteristic UUIDs
type Driver struct {
	Type     string    `yaml:"type"`
	Names    []string  `yaml:"names"`
	Services []Service `yaml:"services"`
}

// A Service allow us to map known service UUIDs to GATT tx/rx characteristic UUIDs
type Service struct {
	// TODO: We might have devices that don't announce a localname, but we know how to talk to.
	// in that case we'll need to one-by-one connect to devices, discover services, and see if the
	// discovered service is something we know how to talk to, then map that back to the appropriate
	// `Implmenetation.`
	ID string `yaml:"id"`
	Tx string `yaml:"tx"`
	Rx string `yaml:"rx"`
}

// GetServicesForIdentifier will return a slice of services know that we can support for this device type
func (c Config) GetServicesForIdentifier(identifier Identifier) []Service {
	services := []Service{}
	found, deviceType := c.GetTypeFromIdentifier(identifier)
	if !found {
		return nil
	}
	for _, driver := range c.Drivers {
		if driver.Type == deviceType {
			for _, service := range driver.Services {
				services = append(services, service)
			}
		}
	}
	return services
}

// GetTypeFromIdentifier will return the driver type we'll use to handle the device identified with
// identifier
func (c Config) GetTypeFromIdentifier(identifier Identifier) (bool, string) {
	for _, driver := range c.Drivers {
		for _, regex := range driver.Names {
			matched, err := regexp.Match(regex, []byte(identifier.Name))
			if err != nil {
				log.Error().Msgf("couldn't execute regex [%s]; skipping: %s", regex, err)
			}
			if matched {
				return true, driver.Type
			}
		}

	}
	return false, ""
}

// TestConfig just generates a test config while we poke around
func TestConfig() Config {
	config := Config{
		Drivers: []Driver{
			{
				Type:  "lovense",
				Names: []string{"LVS-.*", "LOVE-.*"},
				Services: []Service{
					{
						ID: "5a300001-0023-4bd4-bbd5-a6920e4c5653",
						Tx: "5a300002-0023-4bd4-bbd5-a6920e4c5653",
						Rx: "5a300002-0023-4bd4-bbd5-a6920e4c5653",
					},
				},
			},
			{
				Type:  "wevibe",
				Names: []string{"Ditto*"},
				Services: []Service{
					{
						ID: "f000bb03-0451-4000-b000-000000000000",
						Tx: "f000c000-0451-4000-b000-000000000000",
						Rx: "f000b000-0451-4000-b000-000000000000",
					},
				},
			},
			{
				Type:  "pearl21",
				Names: []string{"Pearl2.1"},
				Services: []Service{
					{
						ID: "00001900-0000-1000-8000-00805f9b34fb",
						Tx: "19020000-0000-0000-0000-000000000000",
						Rx: "19030000-0000-0000-0000-000000000000",
					},
				},
			},
		},
	}
	return config
}
