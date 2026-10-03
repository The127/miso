package builder

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// ErrNoShell is what Shell answers when its requests do not end with a shell.
var ErrNoShell = errors.New("the requests do not end with a shell")

// Shell asks the agent every request but the last as Ask does, and the last,
// a shell, on the terminal of the user, whose In is what the user types and
// whose out shows what the shell prints. It answers the code the shell exited
// with. Reading In goes on until it ends, so a caller that keeps running
// closes it.
func Shell(ctx context.Context, vm VM, dial Dial, agent string, requests []build.Request, files Files, term protocol.Terminal, out io.Writer) (int, error) {
	if len(requests) == 0 {
		return 0, ErrNoShell
	}

	last := requests[len(requests)-1]
	shell, isShell := last.Message.(protocol.Shell)
	if !isShell {
		return 0, ErrNoShell
	}

	if err := Ask(ctx, vm, dial, agent, requests[:len(requests)-1], files, nil, out); err != nil {
		return 0, err
	}

	conn, err := connect(ctx, vm, dial, time.After(patienceFor(vm)))
	if err != nil {
		return 0, err
	}

	closeOnCancel := context.AfterFunc(ctx, func() { _ = conn.Close() })
	code, err := protocol.New(agent, conn, conn).AskShell(shell, term, out)
	closeOnCancel()
	_ = conn.Close()

	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if err != nil {
		return 0, last.Failed(failed(vm, err))
	}

	return code, nil
}
