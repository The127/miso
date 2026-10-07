package firmware

// Firmware is a UEFI firmware without Secure Boot, which needs no SMM in
// QEMU.
type Firmware struct {
	Code []byte

	// what the firmware starts from, each boot needs a copy of its own
	Vars []byte
}
