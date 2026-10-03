package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

// openPty makes a terminal and answers its two ends. It has a devpts of its
// own for it in the scratch directory, because the machine's /dev may have
// none, and drops it again, as the ends stay open without it.
func openPty(scratch string) (master, slave *os.File, err error) {
	dir, err := os.MkdirTemp(scratch, "pts")
	if err != nil {
		return nil, nil, err
	}

	defer func() { _ = os.Remove(dir) }()

	if err := syscall.Mount("devpts", dir, "devpts", syscall.MS_NOSUID|syscall.MS_NOEXEC, "newinstance,ptmxmode=0666,mode=0620"); err != nil {
		return nil, nil, fmt.Errorf("mount devpts: %w", err)
	}

	defer func() { _ = syscall.Unmount(dir, syscall.MNT_DETACH) }()

	master, err = os.OpenFile(filepath.Join(dir, "ptmx"), os.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}

	fd := int(master.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		_ = master.Close()

		return nil, nil, err
	}

	number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		_ = master.Close()

		return nil, nil, err
	}

	slave, err = os.OpenFile(filepath.Join(dir, strconv.Itoa(number)), os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		_ = master.Close()

		return nil, nil, err
	}

	return master, slave, nil
}
