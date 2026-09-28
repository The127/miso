package firmware

// Firmware is OVMF without Secure Boot, which needs no SMM in QEMU.
type Firmware struct {
	Code []byte

	// what the firmware starts from, each boot needs a copy of its own
	Vars []byte
}
