package agent

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// helperName is the name the agent starts itself under to run a command.
const helperName = "miso-run"

// Helper sets up a run and becomes its shell when the agent was started as
// the helper of a run, and returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	// the agent reads why a run could not start from here, and the start of
	// the shell closes it
	failed := os.NewFile(3, "failed")
	syscall.CloseOnExec(3)
	root, command := os.Args[1], os.Args[2]
	// returns only when the shell did not start
	err := helper(root, command)
	_, _ = fmt.Fprint(failed, err)
	os.Exit(1)
}

func helper(root, command string) error {
	if err := enterRoot(root); err != nil {
		return err
	}

	for _, m := range mounts {
		if err := m.mount(); err != nil {
			return err
		}
	}

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

	// a fixed name, because packages write it into what they install, and
	// every hosts file knows this one
	if err := syscall.Sethostname([]byte("localhost")); err != nil {
		return fmt.Errorf("name the run: %w", err)
	}

	err = syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", command}, os.Environ()) //nolint:gosec // running what the build file says is what a RUN is

	return fmt.Errorf("run /bin/sh: %w", err)
}
