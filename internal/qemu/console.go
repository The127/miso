package qemu

// console puts the machine's serial port, where its kernel writes with
// console=ttyS0, on QEMU's standard output.
func console() []string {
	return []string{"-serial", "stdio"}
}
