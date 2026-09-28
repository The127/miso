package vsockns

import (
	"encoding/json"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// Program is one started inside the namespace.
type Program struct {
	// where the helper tells how the program ended. A file rather than a
	// number, so that a use after Wait closed it fails instead of reaching
	// whatever took the number since
	report *os.File
}

// Start starts a program inside the namespace, writing to stdout, and
// returns while it runs.
func (n *Namespace) Start(args []string, stdout *os.File) (*Program, error) {
	argument, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	_, _, err = n.ask(startQuestion+" "+string(argument), int(stdout.Fd()), pair[1])
	_ = unix.Close(pair[1])

	if err != nil {
		_ = unix.Close(pair[0])

		return nil, err
	}

	return &Program{report: os.NewFile(uintptr(pair[0]), "report")}, nil
}

// Wait waits until the program ended and gives its exit code.
func (p *Program) Wait() (int, error) {
	defer func() { _ = p.report.Close() }()

	var said string
	err := p.onReport(func(report int) error {
		var err error
		said, _, err = answerOn(report)

		return err
	})
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(said)
}

// Kill stops the program. Wait then tells that it was stopped.
func (p *Program) Kill() error {
	return p.onReport(func(report int) error {
		return say(report, killWord)
	})
}

// onReport does something with the report's fd while it is still open.
func (p *Program) onReport(do func(report int) error) error {
	raw, err := p.report.SyscallConn()
	if err != nil {
		return err
	}

	var failed error
	if err := raw.Control(func(fd uintptr) { failed = do(int(fd)) }); err != nil {
		return err
	}

	return failed
}
