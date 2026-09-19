package boot

import (
	"fmt"
	"syscall"
)

// PowerOff writes everything out and powers the VM off. It never returns,
// because the kernel panics when init does.
func PowerOff() {
	syscall.Sync()

	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		fmt.Println(err)
	}

	select {}
}
