package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// Run runs a command on top of layers and keeps what it writes as the layer
// of the key.
func (a *Agent) Run(ctx context.Context, run protocol.Run, out io.Writer) (int, error) {
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
	// overlay wants its work directory on the file system of the upper one
	scratch, err := a.layers.Scratch()
	if err != nil {
		return 0, err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	root := filepath.Join(scratch, "root")
	overlayWork := filepath.Join(scratch, "work")
	for _, dir := range []string{root, overlayWork} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return 0, err
		}
	}

	// overlay takes the top layer first
	lowers := make([]string, 0, len(run.Layers))
	for _, key := range slices.Backward(run.Layers) {
		lowers = append(lowers, a.layers.Path(key))
	}

	// the run's mount points live below every layer, so a layer holds only
	// what its command wrote
	floor, err := sandbox.Floor(scratch)
	if err != nil {
		return 0, err
	}

	lowers = append(lowers, floor)

	if err := mountOverlay(root, lowers, upper, overlayWork); err != nil {
		return 0, err
	}

	// removing scratch must never reach into the root, so it goes first
	defer func() { _ = syscall.Unmount(root, syscall.MNT_DETACH) }()

	code, err := sandbox.Run(ctx, root, run, out)
	if err != nil || code != 0 {
		return code, err
	}

	// not detached, a busy root means something still writes into the layer
	if err := syscall.Unmount(root, 0); err != nil {
		return 0, err
	}

	return 0, nil
}
