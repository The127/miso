package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/protocol"
)

// Fetch sends the host the disk kept as the output of a key.
func (a *Agent) Fetch(_ context.Context, request protocol.Fetch, pieces protocol.Pieces, _ io.Writer) error {
	disk, err := os.Open(filepath.Join(a.layers.Path(request.Key), "disk.raw"))
	if err != nil {
		return err
	}

	defer func() { _ = disk.Close() }()

	info, err := disk.Stat()
	if err != nil {
		return err
	}

	if err := pieces.Length(info.Size()); err != nil {
		return err
	}

	return pieces.Piece(protocol.Piece{Size: info.Size()}, disk)
}
