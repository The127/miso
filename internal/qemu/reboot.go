package qemu

// noReboot ends QEMU when the machine reboots, which is how a kernel that
// panics with panic=-1 stops, so the host sees the machine is gone.
func noReboot() []string {
	return []string{"-no-reboot"}
}
