## Buttwork

Buttwork provides a mechanism for identifying, connecting to, and controlling, a number of BTLE-enabled vibrating devices.

## Supported Devices

Supported protocols:
* Lovense
* Wevibe (we don't suport the 8bit protocol yet)
* Pearl2.1

## Installation

`go get github.com/psychobummer/buttwork`

NOTE: there are some issues with the way the BTLE package used by `x/bluetooth` is bundled, which makes `go mod` very unhappy; if you get dependency errors when `go mod` tries to resolve them, just `go get tinygo.org/x/bluetooth` and you should be good to go.

## Usage

Usage is pretty straight forward. Scan for BTLE devices, get back a list of devices, connect to the one you want, and then interact with it.

```golang
// Initalize a new discovery agent with the stock device mappings
discovery, err := device.NewBLEDiscovery(device.TestConfig())
if err != nil {
	panic(err)
}

// Scan BTLE announcements for two seconds.
// NOTE: If you'd prefer, you can call `ScanBackground() (<-chan Identifier, <-chan error)`
//       which will return devices as they're found; when you find what you want, just call StopScan()
identifiers, err := discovery.ScanOnce(2 * time.Second)
if err != nil {
	panic(err)
}

// Filter the list of identified devices for any matching the name LVS-*
filteredIdentifiers := identifiers.FilterPrefix("LVS")
if len(filteredIdentifiers) == 0 {
	panic("no supported devices found")
}

// Connect to the first device found.
thisDevice, err := discovery.Connect(filteredIdentifiers[0])
if err != nil {
    panic(err)
}

// Vibrate the device at 20, 48, 60, 80, 100% intensitiy for 2seconds, then disconnect.
// NOTE: All devices have different internal intensity scales; we handle the normalization
//       automatically. 
levels := []int{20, 40, 60, 80, 100}
for level := range levels {
    thisDevice.Vibrate(level)
    <-time.After(2*time.Second)
}
thisDevice.Disconnect()
```

A full example can be found [here](https://github.com/psychobummer/pbrelay-subscriber/blob/master/cmd/midibtle.go)
## FAQ

### Latency on macOS

We've observed sporadic BTLE latencies when macOS has iOS handoff enabled. We don't know enough about the implementation of handoff to speak to the root cause, but disabling it seemed to address the problem.

### Panic on 32bit UUID

```
panic: invalid UUID string: 4C421900
```

This package uses `tinygo.org/x/bluetooth` to handle the underlying BTLE subsystems. On macOS, `x/bluetooth` uses `github.com/JuulLabs-OSS/cbgo` to interface with the macOS's `CoreBluetooth`. `cbgo` handles 16 and 128-bit BTLE UUIDs, but will panic if it encounters a 32-bit UUID.

We've opened a bug and supplied a patch upstream, but it looks like the project might be abandoned. The `go.mod` in this package should pin you to _our_ fork which upscales 32-bit UUIDs to 128-bit, but if for some reason that didn't work out for you, simply install our fork or re-pin it in `go.mod`:

```
replace github.com/JuulLabs-OSS/cbgo => github.com/psychobummer/cbgo v0.0.3-0.20210404042823-cafd0158617d
```

The only device we've found so far which advertises 32-bit UUIDs is the `Pearl2.1`.

### SIG TRAP on macOS BigSur

Starting with BigSur, macOS requires that unsigned applications are explicitly granted access to Bluetooth. If scanning for BTLE devices causes the program to crash with `sigtrap`, you'll need to grant your binary, iterm, whatever, access to Bluetooth. This can be done through `AppleMenu -> System Preferences -> Security & Privacy -> Select Bluetooth from the left column -> Add whatever on the right column`.

## Greetz

We'd like to acknowledge the hard work done by [these folks](https://buttplug.io/). We'd initially tried to implement our system atop of their Rust FFI, but ran into stability issues. Our protocol implementations are 98% cribbed from theirs, though, so we want to ensure credit is given where it's due. They've done a lot of reverse-engineering work.
