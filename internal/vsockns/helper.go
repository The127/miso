package vsockns

import (
	"fmt"
	"os"
	"runtime"
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
	socketQuestion = "socket"
	deviceQuestion = "device"
)

// Helper holds a vsock namespace when miso was started as its helper, and
// returns at once otherwise.
func Helper() {
	if !isHelper(os.Args, helperConn) {
		return
	}

	// the inner namespace is this thread's alone, so everything inside it
	// is done here
	runtime.LockOSThread()

	devicePath = os.Args[2]

	if err := enter(os.Args[1]); err != nil {
		_ = answerFailed(helperConn, err)
		os.Exit(1)
	}

	_ = answerOK(helperConn, "")
	serve(helperConn)
	os.Exit(0)
}

// isHelper is whether the program was started the way miso starts its
// helper: under the helper's name, with the paths the helper works on and
// its channel to miso. A name alone is any program's to take, and it would
// then write into whatever file it names.
func isHelper(args []string, conn int) bool {
	if len(args) < 3 || args[0] != helperName {
		return false
	}

	kind, err := unix.GetsockoptInt(conn, unix.SOL_SOCKET, unix.SO_TYPE)

	return err == nil && kind == unix.SOCK_SEQPACKET
}

// enter makes the namespaces started from here local, and moves this thread
// into one. The helper's own namespace takes the host's mode, which only
// root may change, and only once.
func enter(childMode string) error {
	//nolint:gosec // miso names the path when it starts its own helper
	if err := os.WriteFile(childMode, []byte("local"), 0); err != nil {
		return err
	}

	if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
		return err
	}

	// a kernel may take the write and still not make the namespace local
	now, _, err := mode("", nil)
	if err != nil {
		return err
	}

	if now != "local" {
		return fmt.Errorf("the namespace's vsock mode is %s, not local", now)
	}

	return nil
}

// handler answers one kind of question, with the files it hands over. The
// files miso handed with the question are only lent to it.
type handler func(argument string, files []int) (string, []int, error)

var handlers = map[string]handler{
	modeQuestion:   mode,
	socketQuestion: socket,
	deviceQuestion: device,
}

// serve answers miso until it hangs up.
func serve(conn int) {
	for {
		asked, received, err := hear(conn)
		if err != nil {
			return
		}

		respond(conn, asked, received)

		for _, file := range received {
			_ = unix.Close(file)
		}
	}
}

// respond answers one question, with the files miso handed with it.
func respond(conn int, asked string, received []int) {
	question, argument, _ := strings.Cut(asked, " ")
	handle, known := handlers[question]
	if !known {
		// miso would wait for an answer forever
		_ = answerFailed(conn, fmt.Errorf("unknown question %s", question))

		return
	}

	text, files, err := handle(argument, received)
	if err != nil {
		_ = answerFailed(conn, err)

		return
	}

	_ = answerOK(conn, text, files...)

	// miso holds its own copies now
	for _, file := range files {
		_ = unix.Close(file)
	}
}

// mode is the vsock mode of the namespace the helper works in.
func mode(string, []int) (string, []int, error) {
	said, err := os.ReadFile("/proc/sys/net/vsock/ns_mode")
	if err != nil {
		return "", nil, err
	}

	return strings.TrimSpace(string(said)), nil, nil
}

// socket is a fresh vsock socket made inside.
func socket(string, []int) (string, []int, error) {
	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return "", nil, err
	}

	return "", []int{fd}, nil
}

// device is the host's vhost-vsock device opened inside.
func device(string, []int) (string, []int, error) {
	fd, err := unix.Open(devicePath, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return "", nil, fmt.Errorf("the host's vsock device %s needs the vhost_vsock module and access for this user: %w", devicePath, err)
	}

	return "", []int{fd}, nil
}
