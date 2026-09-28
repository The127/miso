package vsockns

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// helperName is the name miso starts itself under to hold a namespace.
const helperName = "miso-vsockns"

// helperConn is the helper's end of its channel to miso, the first file
// miso hands it, since those start after stdin, stdout and stderr.
const helperConn = 3

const (
	modeQuestion   = "mode"
	listenQuestion = "listen"
	socketQuestion = "socket"
)

// Helper holds a vsock namespace when miso was started as its helper, and
// returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	// the inner namespace is this thread's alone, so everything inside it
	// is done here
	runtime.LockOSThread()

	if err := enter(os.Args[1]); err != nil {
		_ = answerFailed(helperConn, err)
		os.Exit(1)
	}

	_ = answerOK(helperConn, "")
	serve(helperConn)
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

// handler answers one kind of question, with the files it hands over.
type handler func(argument string) (string, []int, error)

var handlers = map[string]handler{
	modeQuestion:   mode,
	listenQuestion: listenOn,
	socketQuestion: socket,
}

// serve answers miso until it hangs up.
func serve(conn int) {
	for {
		asked, _, err := hear(conn)
		if err != nil {
			return
		}

		question, argument, _ := strings.Cut(asked, " ")
		handle, known := handlers[question]
		if !known {
			// miso would wait for an answer forever
			_ = answerFailed(conn, fmt.Errorf("unknown question %s", question))

			continue
		}

		text, files, err := handle(argument)
		if err != nil {
			_ = answerFailed(conn, err)

			continue
		}

		_ = answerOK(conn, text, files...)

		// miso holds its own copies now
		for _, file := range files {
			_ = unix.Close(file)
		}
	}
}

// mode is the vsock mode of the namespace the helper works in.
func mode(string) (string, []int, error) {
	said, err := os.ReadFile("/proc/sys/net/vsock/ns_mode")
	if err != nil {
		return "", nil, err
	}

	return strings.TrimSpace(string(said)), nil, nil
}

// listenOn is a listening vsock socket made inside, on the port asked for.
func listenOn(argument string) (string, []int, error) {
	port, err := strconv.ParseUint(argument, 10, 32)
	if err != nil {
		return "", nil, err
	}

	listener, err := listen(uint32(port))
	if err != nil {
		return "", nil, err
	}

	return "", []int{listener}, nil
}

// socket is a fresh vsock socket made inside.
func socket(string) (string, []int, error) {
	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return "", nil, err
	}

	return "", []int{fd}, nil
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
