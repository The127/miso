package agent

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
)

// Copy puts what the host sends on top of layers and keeps it as the layer
// of the key.
func (a *Agent) Copy(_ context.Context, request protocol.Copy, entries protocol.Entries, _ io.Writer) error {
	work, err := a.layers.Begin(request.Key)
	if err != nil {
		return err
	}

	err = a.overlaid(request.Layers, work.Dir(), func(root string) (bool, error) {
		return true, placeAll(place.Open(root), request.Destination, entries)
	})
	if err != nil {
		_ = work.Discard()

		return err
	}

	return work.Finish()
}

// placeAll puts every entry below the destination, the source itself at
// the destination.
func placeAll(image *place.Root, destination string, entries protocol.Entries) error {
	for {
		entry, content, err := entries.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		if err := placeOne(image, filepath.Join(destination, entry.Path), entry, content); err != nil {
			return err
		}
	}
}

// placeOne puts an entry at a path of the image as what it is.
func placeOne(image *place.Root, path string, entry protocol.Entry, content io.Reader) error {
	switch entry.Kind {
	case "directory":
		return image.Directory(path, entry.Mode)
	case "link":
		return image.Link(path, entry.Target)
	default:
		return image.File(path, entry.Mode, content)
	}
}
