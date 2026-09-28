package vsockns

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// helperName is the name miso starts itself under to hold a namespace.
const helperName = "miso-vsockns"

const (
	readyWord    = "ready"
	modeQuestion = "mode"
)

// Helper holds a vsock namespace when miso was started as its helper, and
// returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	conn := os.NewFile(3, "vsockns")

	// the inner namespace is this thread's alone, so everything inside it
	// is done here
	runtime.LockOSThread()

	if err := enter(os.Args[1]); err != nil {
		_, _ = fmt.Fprintln(conn, err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintln(conn, readyWord)
	serve(conn)
	os.Exit(0)
}

// enter makes the namespaces started from here local, and moves this thread
// into one. The helper's own namespace takes the host's mode, which only
// root may change, and only once.
func enter(childMode string) error {
	//nolint:gosec // miso names the path when it starts its own helper
	if err := os.WriteFile(childMode, []byte("local"), 0); err != nil {
		return err
	}

	return unix.Unshare(unix.CLONE_NEWNET)
}

// serve answers miso until it hangs up.
func serve(conn *os.File) {
	asked := bufio.NewScanner(conn)
	for asked.Scan() {
		if strings.TrimSpace(asked.Text()) == modeQuestion {
			// a helper that cannot tell ends, and miso hears no answer
			mode, err := os.ReadFile("/proc/sys/net/vsock/ns_mode")
			if err != nil {
				return
			}

			_, _ = fmt.Fprintln(conn, strings.TrimSpace(string(mode)))
		}
	}
}
