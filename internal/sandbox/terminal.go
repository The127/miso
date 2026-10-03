package sandbox

import (
	"context"
	"io"
	"os"
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

	if err := resize(master, shell.Rows, shell.Cols); err != nil {
		return 0, err
	}

	go func() { _, _ = io.Copy(master, term.In) }()

	printed := make(chan struct{})
	go func() {
		defer close(printed)

		// ends with EIO once no shell holds the terminal any more
		_, _ = io.Copy(out, master)
	}()

	go term.Follow(printed, func(size protocol.Resize) bool {
		// a terminal that is gone has no size to keep
		_ = resize(master, size.Rows, size.Cols)

		return true
	})

	code, err := start(ctx, root, argsOf(shell), envOf(shell), shell.Network, streams{terminal: slave})
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

// resize sets the size of the terminal the master belongs to. A size of
// nothing is what a terminal that does not know its size reports, and is
// no size.
func resize(master *os.File, rows, cols uint16) error {
	if rows == 0 && cols == 0 {
		return nil
	}

	return unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: rows, Col: cols})
}

// argsOf are the arguments of the sh that is a shell, or that runs the words
// of a command as they are, looking its program up on the PATH of the root.
// The command is a child of the sh and not the sh itself, because the sh is
// the init of the PID namespace, which the kernel keeps from being stopped
// by the signals of a terminal. The exit keeps the sh from replacing itself
// with the command.
func argsOf(shell protocol.Shell) []string {
	if len(shell.Command) == 0 {
		return []string{"-i"}
	}

	return append([]string{"-c", `"$@"; exit $?`, "miso-shell"}, shell.Command...)
}
