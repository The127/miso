package agent

import (
	"path/filepath"
	"strings"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
)

// target is where an entry of a copy from the build context lands in the
// image.
func target(image *place.Root, request protocol.Copy, entry protocol.Entry) (string, error) {
	return landing(image, request.Destination, request.Sources[entry.Source], entry.Path, entry.Kind == "directory")
}

// landing is where something at a path below a source lands in the image.
// What is in a directory source goes below the destination, the directory
// itself is the destination. A file or link source goes under its own name
// into a destination that ends in a slash or is a directory of the image
// already, and is the destination otherwise.
func landing(image *place.Root, destination, source, path string, directory bool) (string, error) {
	if path != "." || directory {
		return filepath.Join(destination, path), nil
	}

	into := strings.HasSuffix(destination, "/")
	if !into {
		var err error
		if into, err = image.IsDirectory(destination); err != nil {
			return "", err
		}
	}

	if into {
		// a source is walked by its clean path, so motd/. is motd
		return filepath.Join(destination, filepath.Base(filepath.Clean(source))), nil
	}

	return destination, nil
}
