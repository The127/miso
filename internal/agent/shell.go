package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// Shell runs an interactive shell on top of layers, with everything it
// writes thrown away once it ends.
func (a *Agent) Shell(ctx context.Context, shell protocol.Shell, in io.Reader, out io.Writer) (int, error) {
	scratch, err := a.layers.Scratch()
	if err != nil {
		return 0, err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	upper := filepath.Join(scratch, "upper")
	if err := os.Mkdir(upper, 0o700); err != nil {
		return 0, err
	}

	code := 0
	err = a.overlaid(shell.Layers, sandbox.Floor, ceilingFor(shell.Network), upper, func(root string) (bool, error) {
		var err error
		code, err = sandbox.Shell(ctx, root, scratch, shell, in, out)

		return false, err
	})

	return code, err
}
