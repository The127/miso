package initramfs

// Emulator is the QEMU that runs the programs of an architecture other than
// the builder VM's.
type Emulator struct {
	Arch    string
	Program []byte
}
