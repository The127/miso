package agent

import (
	"context"
	"errors"
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Fetch fails, the agent sends no disks yet.
func (a *Agent) Fetch(context.Context, protocol.Fetch, protocol.Pieces, io.Writer) error {
	return errors.New("the agent sends no disks yet")
}
