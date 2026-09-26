package qemu

// console puts the machine's serial port, where its kernel writes with
// console=ttyS0, on QEMU's standard output. QEMU opens a window on a host
// with a desktop, and on one without none, so the window is always off.
func console() []string {
	return []string{"-serial", "stdio", "-display", "none"}
}
