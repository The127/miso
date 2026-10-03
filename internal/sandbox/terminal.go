package sandbox

import (
	"context"
	"io"
	"os/exec"

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

	go func() { _, _ = io.Copy(master, in) }()

	printed := make(chan struct{})
	go func() {
		defer close(printed)

		// ends with EIO once no shell holds the terminal any more
		_, _ = io.Copy(out, master)
	}()

	code, err := start(ctx, root, []string{"-i"}, shell.Env, shell.Network, func(cmd *exec.Cmd) {
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
