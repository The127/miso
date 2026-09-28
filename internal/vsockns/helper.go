package vsockns

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

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
	runQuestion    = "run"
	startQuestion  = "start"

	// what miso sends on a program's report to stop it
	killWord = "kill"
)

// Helper holds a vsock namespace when miso was started as its helper, and
// returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	// what the helper runs must not be able to talk to miso as the helper,
	// nor hold what miso's own caller left open for it
	syscall.CloseOnExec(helperConn)

	if err := unix.CloseRange(helperConn+1, math.MaxUint32, unix.CLOSE_RANGE_CLOEXEC); err != nil {
		_ = answerFailed(helperConn, err)
		os.Exit(1)
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

// handler answers one kind of question, with the files it hands over. The
// files miso handed with the question are only lent to it.
type handler func(argument string, files []int) (string, []int, error)

var handlers = map[string]handler{
	modeQuestion:   mode,
	socketQuestion: socket,
	runQuestion:    run,
	startQuestion:  start,
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

// run runs a program to its end, from this thread, so in the namespace, with
// the file miso handed as its stdout.
func run(argument string, files []int) (string, []int, error) {
	if len(files) != 1 {
		return "", nil, fmt.Errorf("a program needs its stdout, it got %d files", len(files))
	}

	// a copy of its own, since the file is only lent, and one the program
	// does not inherit beside its stdout
	own, err := unix.FcntlInt(uintptr(files[0]), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return "", nil, err
	}

	stdout := os.NewFile(uintptr(own), "stdout")
	defer func() { _ = stdout.Close() }()

	var args []string
	if err := json.Unmarshal([]byte(argument), &args); err != nil {
		return "", nil, err
	}

	if len(args) == 0 {
		return "", nil, errors.New("a program needs a name")
	}

	//nolint:gosec // miso names the program it runs in its own namespace
	program := exec.Command(args[0], args[1:]...)
	program.Stdout = stdout

	return "", nil, program.Run()
}

// start starts a program from this thread, so in the namespace, with the
// first file miso lent as its stdout, and tells on the second how it ended.
func start(argument string, files []int) (string, []int, error) {
	if len(files) != 2 {
		return "", nil, fmt.Errorf("a program needs its stdout and its report, it got %d files", len(files))
	}

	var args []string
	if err := json.Unmarshal([]byte(argument), &args); err != nil {
		return "", nil, err
	}

	if len(args) == 0 {
		return "", nil, errors.New("a program needs a name")
	}

	// copies of their own, since the files are only lent, and ones the
	// program does not inherit
	own, err := unix.FcntlInt(uintptr(files[0]), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return "", nil, err
	}

	stdout := os.NewFile(uintptr(own), "stdout")
	defer func() { _ = stdout.Close() }()

	report, err := unix.FcntlInt(uintptr(files[1]), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return "", nil, err
	}

	//nolint:gosec // miso names the program it runs in its own namespace
	program := exec.Command(args[0], args[1:]...)
	program.Stdout = stdout

	if err := program.Start(); err != nil {
		_ = unix.Close(report)

		return "", nil, err
	}

	// the waiter alone closes the report, once it has woken the listener
	// and the listener is done, so neither uses the number after it went
	listening := make(chan struct{})
	go func() {
		defer close(listening)

		stopOnWord(program, report)
	}()
	go tellEnd(program, report, listening)

	return "", nil, nil
}

// stopOnWord kills the program once miso says so, or its report ends. It
// listens on after a kill, until the waiter has told how the program ended.
func stopOnWord(program *exec.Cmd, report int) {
	for {
		said, _, err := hear(report)
		if err != nil {
			// miso is gone, or the waiter is done, and a program that ended
			// is not killed
			_ = program.Process.Kill()

			return
		}

		if said == killWord {
			_ = program.Process.Kill()
		}
	}
}

// tellEnd tells miso how the program ended, then closes the report once
// the listener is done with it.
func tellEnd(program *exec.Cmd, report int, listening <-chan struct{}) {
	defer func() {
		_ = unix.Shutdown(report, unix.SHUT_RDWR)
		<-listening
		_ = unix.Close(report)
	}()

	err := program.Wait()
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		_ = answerFailed(report, err)

		return
	}

	if status, ok := program.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		_ = answerFailed(report, fmt.Errorf("the program was %s", program.ProcessState))

		return
	}

	_ = answerOK(report, strconv.Itoa(program.ProcessState.ExitCode()))
}
