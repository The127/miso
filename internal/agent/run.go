package agent

import (
	"context"
	"io"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// Run runs a command on top of layers and keeps what it writes as the layer
// of the key. A key whose layer is there already has run before.
func (a *Agent) Run(ctx context.Context, run protocol.Run, out io.Writer) (int, error) {
	there, err := a.found(run.Key, run.Layers)
	if err != nil || there {
		return 0, err
	}

	work, err := a.layers.Begin(run.Key)
	if err != nil {
		return 0, err
	}

	code, err := a.runOn(ctx, work.Dir(), run, out)
	if err != nil || code != 0 {
		_ = work.Discard()

		return code, err
	}

	return 0, work.Finish()
}

// runOn runs a command on top of layers with what it writes going into a
// directory, which is no longer mounted once it returns.
func (a *Agent) runOn(ctx context.Context, upper string, run protocol.Run, out io.Writer) (int, error) {
	code := 0
	var ceiling func(scratch string, below []string) (string, error)
	if run.Network != nil {
		ceiling = func(scratch string, below []string) (string, error) {
			return sandbox.Ceiling(scratch, below, run.Network)
		}
	}

	// the run's mount points live below every layer, so a layer holds only
	// what its command wrote
	err := a.overlaid(run.Layers, sandbox.Floor, ceiling, upper, func(root string) (bool, error) {
		var err error
		code, err = sandbox.Run(ctx, root, run, out)

		return code == 0, err
	})

	return code, err
}
