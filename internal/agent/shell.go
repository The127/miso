package agent

import (
	"context"
	"errors"
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Shell is not there yet.
func (a *Agent) Shell(context.Context, protocol.Shell, io.Reader, io.Writer) (int, error) {
	return 0, errors.ErrUnsupported
}
