package main

import (
	"log"
	"time"

	"github.com/psychobummer/buttwork/device"
)

// Example client usage.
// * Connect to a device with a prefix of "LVS"
// * Tell it to vibrate for 5 seconds

func main() {
	discovery, err := device.NewBLEDiscovery()
	if err != nil {
		log.Fatal(err)
	}

	identifiers, err := discovery.ScanOnce(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	filteredIdentifiers := identifiers.FilterPrefix("LVS")
	if len(filteredIdentifiers) == 0 {
		log.Fatal("No compatible devices found")
	}

	device, err := discovery.Connect(filteredIdentifiers[0])
	if err != nil {
		log.Fatalf("Couldn't connect to device %s", filteredIdentifiers[0].Address)
	}

	_, err = device.Vibrate(5)
	if err != nil {
		log.Fatal("Couldn't issue vibrate command")
	}
	<-time.After(5 * time.Second)
	_, err = device.Vibrate(0)
	if err != nil {
		log.Fatal("Couldn't issue vibrate command")
	}

	if err = device.Disconnect(); err != nil {
		log.Fatal("Couldn't disconnect from device")
	}

}
