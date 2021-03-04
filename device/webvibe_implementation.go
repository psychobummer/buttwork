package device

// WevibeImplementation provides a concrete implmentation of the single-motor webvibe protocol
type WevibeImplementation struct{}

// Init is a NOOP for wevibe devices
func (l WevibeImplementation) Init() error {
	return nil
}

// Name returns the implementation name
func (l WevibeImplementation) Name() string {
	return "wevibe"
}

// VibrateCommand will vibrate a single-motor webvibe device.
// Wevibe appears to take an 8-byte sequence, though i'm not entirely sure
// if this is 100% correct. finding conflicting information.
//
// 0x0f, <vibe pattern>, 0x00, <intensity>, 0x00, 0x03, 0x00, 0x00
//
// notes:
// setting the <intensity> bits to 0x00 doesn't appear to silence the device,
// but changing the <vibe pattern> bits to 0x00 does; an intensity of 0x00 is maybe
// used to represent the lowest possible value instead of "off", and a pattern of
// 0x00 is off?
func (l WevibeImplementation) VibrateCommand(level uint8) []byte {
	if level <= 0 {
		return []byte{0x0f, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00}
	}
	deviceMax := uint8(15)
	normalized := normalize(level, deviceMax)
	command := []byte{0x0f, 0x03, 0x00, normalized, 0x00, 0x03, 0x00, 0x00}
	return []byte(command)
}
