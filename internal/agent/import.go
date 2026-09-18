package agent

import (
	"context"
	"io"
	"os"
	"syscall"

	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/tree"
)

// Agent does what the host asks, keeping layers in one directory and
// mounting base images below another.
type Agent struct {
	layers  *layer.Store
	scratch string
}

// New takes the directory the layers live in and one for mount points. It
// touches nothing yet.
func New(layers, scratch string) *Agent {
	return &Agent{layers: layer.Open(layers), scratch: scratch}
}

// Import keeps the root file system of a base image as the layer of a key.
func (a *Agent) Import(_ context.Context, request protocol.Import, _ io.Writer) error {
	there, err := a.layers.Has(request.Key)
	if err != nil || there {
		return err
	}

	work, err := a.layers.Begin(request.Key)
	if err != nil {
		return err
	}

	base, err := os.MkdirTemp(a.scratch, "base-")
	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(base) }()

	if err := mountRoot(protocol.Serial(request.Digest), base); err != nil {
		return err
	}

	defer func() { _ = syscall.Unmount(base, syscall.MNT_DETACH) }()

	if err := tree.Copy(base, work.Dir()); err != nil {
		return err
	}

	return work.Finish()
}
