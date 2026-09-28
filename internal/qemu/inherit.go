package qemu

import (
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// inheritNothing marks every file of miso's close-on-exec, so QEMU holds
// only what it is handed. Go keeps a file open across exec that it did not
// open itself, such as one miso's own caller left open. /proc is read
// rather than close_range used, which older kernels lack.
func inheritNothing() error {
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return err
	}

	for _, entry := range fds {
		fd, err := strconv.Atoi(entry.Name())
		if err != nil || fd <= 2 {
			continue
		}

		// the directory read above is gone by now, and every other file
		// may have been closed since
		_, _ = unix.FcntlInt(uintptr(fd), unix.F_SETFD, unix.FD_CLOEXEC)
	}

	return nil
}
