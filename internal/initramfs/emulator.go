package initramfs

// Emulator is the QEMU that runs the programs of an architecture other than
// the builder VM's, and where it is from.
type Emulator struct {
	Arch    string
	From    string
	Program []byte
}
