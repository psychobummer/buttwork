package device

// Pearl21Implementation provides a concrete implmentation of the Pearl21 protocol
type Pearl21Implementation struct{}

// Init is a NOOP for Pearl21 devices
func (l Pearl21Implementation) Init() error {
	return nil
}

// Name returns the implementation name
func (l Pearl21Implementation) Name() string {
	return "pearl21"
}

// VibrateCommand will vibrate a Pearl21 device
func (l Pearl21Implementation) VibrateCommand(level uint8) []byte {
	deviceMax := uint8(100)
	normalized := normalize(level, deviceMax)
	command := []byte{0x01, normalized}
	return command
}
