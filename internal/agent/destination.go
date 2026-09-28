package agent

import (
	"path/filepath"
	"strings"

	"github.com/The127/miso/internal/protocol"
)

// target is where an entry lands in the image. A directory's contents go
// into the destination. A file or link goes into a destination that ends
// in a slash under the name of its source, and is the destination
// otherwise.
func target(request protocol.Copy, entry protocol.Entry) string {
	if entry.Kind != "directory" && strings.HasSuffix(request.Destination, "/") {
		return filepath.Join(request.Destination, filepath.Base(request.Sources[entry.Source]), entry.Path)
	}

	return filepath.Join(request.Destination, entry.Path)
}
