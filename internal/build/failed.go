package build

import (
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// Failed is why the request failed, said at the line of the build file it
// came from and as its author wrote it there.
func (r Request) Failed(err error) error {
	return &imagefile.Error{Line: r.Line, Err: fmt.Errorf("%s: %w", r.Written, err)}
}
