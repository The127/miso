package vsockns

import (
	"encoding/json"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// Program is one started inside the namespace.
type Program struct {
	// where the helper tells how the program ended
	report int
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

	return &Program{report: pair[0]}, nil
}

// Wait waits until the program ended and gives its exit code.
func (p *Program) Wait() (int, error) {
	defer func() { _ = unix.Close(p.report) }()

	said, _, err := answerOn(p.report)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(said)
}
