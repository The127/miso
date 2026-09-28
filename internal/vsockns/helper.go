package vsockns

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// helperName is the name miso starts itself under to hold a namespace.
const helperName = "miso-vsockns"

const (
	modeQuestion   = "mode"
	listenQuestion = "listen"
)

// Helper holds a vsock namespace when miso was started as its helper, and
// returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	const conn = 3

	// the inner namespace is this thread's alone, so everything inside it
	// is done here
	runtime.LockOSThread()

	if err := enter(os.Args[1]); err != nil {
		_ = answerFailed(conn, err)
		os.Exit(1)
	}

	_ = answerOK(conn, "")
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
func serve(conn int) {
	for {
		asked, _, err := hear(conn)
		if err != nil {
			return
		}

		question, argument, _ := strings.Cut(asked, " ")
		switch question {
		case modeQuestion:
			// a helper that cannot tell ends, and miso hears no answer
			mode, err := os.ReadFile("/proc/sys/net/vsock/ns_mode")
			if err != nil {
				return
			}

			_ = answerOK(conn, strings.TrimSpace(string(mode)))
		case listenQuestion:
			answerListen(conn, argument)
		}
	}
}

// answerListen hands miso a listening vsock socket made inside.
func answerListen(conn int, argument string) {
	port, err := strconv.ParseUint(argument, 10, 32)
	if err != nil {
		_ = answerFailed(conn, err)

		return
	}

	listener, err := listen(uint32(port))
	if err != nil {
		_ = answerFailed(conn, err)

		return
	}

	_ = answerOK(conn, "", listener)
	_ = unix.Close(listener)
}

func listen(port uint32) (int, error) {
	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}

	err = unix.Bind(fd, &unix.SockaddrVM{CID: unix.VMADDR_CID_ANY, Port: port})
	if err == nil {
		err = unix.Listen(fd, unix.SOMAXCONN)
	}

	if err != nil {
		_ = unix.Close(fd)

		return -1, err
	}

	return fd, nil
}
