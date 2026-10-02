package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
)

// BootPart keeps the kernel or the initrd of the image's layers as the layer
// of a key. A key whose layer is there already has its part.
func (a *Agent) BootPart(_ context.Context, request protocol.BootPart, _ io.Writer) error {
	if request.Part != protocol.PartKernel && request.Part != protocol.PartInitrd {
		return fmt.Errorf("no part %q can be kept", request.Part)
	}

	there, at, err := a.found(request.Key, request.Layers)
	if err != nil || there {
		return err
	}

	scratch, err := a.layers.Scratch()
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	image, err := a.mountImage(scratch, request.Layers)
	if err != nil {
		return err
	}

	// removing scratch must never reach into the image, so it goes first
	defer image.unmount()

	// links in the image mean places in the image, never in the builder VM
	imageFS := place.Open(image.dir).FS()

	found, err := kernel.Find(imageFS, "")
	if err != nil {
		return err
	}

	path := found.Linux
	if request.Part == protocol.PartInitrd {
		path = found.Initrd
	}

	work, err := a.layers.Begin(request.Key)
	if err != nil {
		return err
	}

	if err := copyPart(imageFS, path, filepath.Join(work.Dir(), outputFile)); err != nil {
		_ = work.Discard()

		return err
	}

	return work.FinishAt(at)
}
