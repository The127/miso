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

// CheckFailed is a check of the request that exited with a code, said at
// its own line with what it printed.
func (r Request) CheckFailed(index, code int, output string) error {
	line, written := imagefile.Written(r.Checks[index])

	return &imagefile.Error{Line: line, Err: fmt.Errorf("%s: exit code %d\n%s", written, code, output)}
}
