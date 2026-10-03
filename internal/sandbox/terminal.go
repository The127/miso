package sandbox

import (
	"context"
	"io"
	"os"
	"os/exec"
	"slices"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
)

// Shell runs an interactive shell in a root on a terminal of its own, which
// reads in and writes out, follows the sizes the user's terminal changes to,
// and answers its exit code. Reading in goes on until it ends, so the
// caller closes it.
func Shell(ctx context.Context, root, scratch string, shell protocol.Shell, term protocol.Terminal, out io.Writer) (int, error) {
	master, slave, err := openPty(scratch)
	if err != nil {
		return 0, err
	}

	defer func() { _ = master.Close() }()

	if shell.Rows != 0 || shell.Cols != 0 {
		if err := resize(master, shell.Rows, shell.Cols); err != nil {
			return 0, err
		}
	}

	go func() { _, _ = io.Copy(master, term.In) }()

	printed := make(chan struct{})
	go func() {
		defer close(printed)

		// ends with EIO once no shell holds the terminal any more
		_, _ = io.Copy(out, master)
	}()

	go func() {
		for {
			select {
			case size := <-term.Resized:
				// a terminal that is gone has no size to keep
				_ = resize(master, size.Rows, size.Cols)
			case <-printed:
				return
			}
		}
	}()

	code, err := start(ctx, root, []string{"-i"}, envOf(shell), shell.Network, func(cmd *exec.Cmd) {
		cmd.Stdin = slave
		cmd.Stdout = slave
		cmd.Stderr = slave
		cmd.SysProcAttr.Setsid = true
		cmd.SysProcAttr.Setctty = true
	})
	_ = slave.Close()
	<-printed

	return code, err
}

// envOf is the environment of the step and the terminal type of the user,
// which wins, because the user is the one at the terminal.
func envOf(shell protocol.Shell) []string {
	if shell.Term == "" {
		return shell.Env
	}

	return append(slices.Clone(shell.Env), "TERM="+shell.Term)
}

// resize sets the size of the terminal the master belongs to.
func resize(master *os.File, rows, cols uint16) error {
	return unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: rows, Col: cols})
}
