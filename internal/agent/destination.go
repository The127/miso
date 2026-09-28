package agent

import (
	"path/filepath"
	"strings"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
)

// target is where an entry lands in the image. What is in a directory
// source goes below the destination, the directory itself is the
// destination. A file or link source goes under its own name into a
// destination that ends in a slash or is a directory of the image already,
// and is the destination otherwise.
func target(image *place.Root, request protocol.Copy, entry protocol.Entry) (string, error) {
	if entry.Path != "." || entry.Kind == "directory" {
		return filepath.Join(request.Destination, entry.Path), nil
	}

	into := strings.HasSuffix(request.Destination, "/")
	if !into {
		var err error
		if into, err = image.IsDirectory(request.Destination); err != nil {
			return "", err
		}
	}

	if into {
		// the host walks a source by its clean path, so motd/. is motd
		return filepath.Join(request.Destination, filepath.Base(filepath.Clean(request.Sources[entry.Source]))), nil
	}

	return request.Destination, nil
}
