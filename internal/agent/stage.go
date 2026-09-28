package agent

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/tree"
)

// copyStage copies the sources of an earlier stage, which its layers hold,
// into the image at root. Nothing comes from the host.
func (a *Agent) copyStage(request protocol.Copy, root string) error {
	scratch, err := a.layers.Scratch()
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	stage := filepath.Join(scratch, "stage")
	if err := os.Mkdir(stage, 0o700); err != nil {
		return err
	}

	lowers, err := a.lowers(request.From)
	if err != nil {
		return err
	}

	// overlay without an upper layer wants two below
	bottom, err := empty(scratch)
	if err != nil {
		return err
	}

	if err := mountReadOnly(stage, append(lowers, bottom)); err != nil {
		return err
	}

	// removing scratch must never reach into the stage, so it goes first
	defer func() { _ = syscall.Unmount(stage, syscall.MNT_DETACH) }()

	image := place.Open(root)
	land := func(source, below string, directory bool) (string, error) {
		return landing(image, request.Destination, source, below, directory)
	}

	return tree.Into(stage, image, request.Sources, land)
}
