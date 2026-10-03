package sandbox

import (
	"context"
	"io"
	"os/exec"
	"slices"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
)

// Shell runs an interactive shell in a root on a terminal of its own, which
// reads in and writes out, and answers its exit code. Reading in goes on
// until it ends, so the caller closes it.
func Shell(ctx context.Context, root, scratch string, shell protocol.Shell, in io.Reader, out io.Writer) (int, error) {
	master, slave, err := openPty(scratch)
	if err != nil {
		return 0, err
	}

	defer func() { _ = master.Close() }()

	if shell.Rows != 0 || shell.Cols != 0 {
		size := &unix.Winsize{Row: shell.Rows, Col: shell.Cols}
		if err := unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, size); err != nil {
			return 0, err
		}
	}

	go func() { _, _ = io.Copy(master, in) }()

	printed := make(chan struct{})
	go func() {
		defer close(printed)

		// ends with EIO once no shell holds the terminal any more
		_, _ = io.Copy(out, master)
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
