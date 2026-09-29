package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
)

// Fetch sends the host the disk kept as the output of a key.
func (a *Agent) Fetch(_ context.Context, request protocol.Fetch, pieces protocol.Pieces, _ io.Writer) error {
	// the tools that wrote the disk run the build's code, so its name may be
	// a link to anything in the builder
	disk, err := os.OpenFile(filepath.Join(a.layers.Path(request.Key), "disk.raw"), os.O_RDONLY|unix.O_NOFOLLOW, 0)
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

	for offset := int64(0); ; {
		data, err := disk.Seek(offset, unix.SEEK_DATA)
		// past the last data, the rest of the disk is a hole
		if errors.Is(err, unix.ENXIO) {
			return nil
		}

		if err != nil {
			return err
		}

		hole, err := disk.Seek(data, unix.SEEK_HOLE)
		if err != nil {
			return err
		}

		piece := protocol.Piece{Offset: data, Size: hole - data}
		if err := pieces.Piece(piece, io.NewSectionReader(disk, data, piece.Size)); err != nil {
			return err
		}

		offset = hole
	}
}
