package main

import (
	"time"

	"github.com/psychobummer/buttwork/device"
	"github.com/rs/zerolog/log"
)

// Example client usage.
// * Connect to a device with a prefix of "LVS"
// * Tell it to vibrate for 5 seconds

func main() {

	config := device.TestConfig()
	//fmt.Printf("%+v", config)

	discovery, err := device.NewBLEDiscovery(config)
	if err != nil {
		log.Fatal().Err(err)
	}

	identifiers, err := discovery.ScanOnce(5 * time.Second)
	if err != nil {
		log.Fatal().Err(err)
	}

	filteredIdentifiers := identifiers.FilterPrefix("LVS")
	if len(filteredIdentifiers) == 0 {
		log.Fatal().Msg("No compatible devices found")
	}

	device, err := discovery.Connect(filteredIdentifiers[0])
	if err != nil {
		log.Fatal().Msgf("Couldn't connect to device %s: %s", filteredIdentifiers[0].Address, err)
	}

	_, err = device.Vibrate(5)
	if err != nil {
		log.Fatal().Msg("Couldn't issue vibrate command")
	}
	<-time.After(5 * time.Second)
	_, err = device.Vibrate(0)
	if err != nil {
		log.Fatal().Msg("Couldn't issue vibrate command")
	}

	if err = device.Disconnect(); err != nil {
		log.Fatal().Msg("Couldn't disconnect from device")
	}

}
