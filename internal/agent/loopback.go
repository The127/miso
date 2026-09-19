package agent

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// upLoopback brings up the loopback of the run's network namespace.
func upLoopback() error {
	// a new network namespace starts with its loopback down
	sock, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(sock) }()

	lo, err := unix.NewIfreq("lo")
	if err != nil {
		return err
	}

	if err := unix.IoctlIfreq(sock, unix.SIOCGIFFLAGS, lo); err != nil {
		return fmt.Errorf("read lo: %w", err)
	}

	lo.SetUint16(lo.Uint16() | unix.IFF_UP)
	if err := unix.IoctlIfreq(sock, unix.SIOCSIFFLAGS, lo); err != nil {
		return fmt.Errorf("bring lo up: %w", err)
	}

	return nil
}
