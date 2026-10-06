package qemu

// console puts the machine's serial port, where its kernel writes with
// console= and its SerialConsole, on QEMU's standard output. QEMU opens a
// window on a host with a desktop, and on one without none, so the window is
// always off.
func console() []string {
	return []string{"-serial", "stdio", "-display", "none"}
}

// SerialConsole is what Linux calls the serial port of a machine of an
// architecture. The virt board has a PL011 instead of a 16550 UART.
func SerialConsole(arch string) string {
	if arch == arm64 {
		return "ttyAMA0"
	}

	return "ttyS0"
}
