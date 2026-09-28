package agent

import (
	"context"
	"errors"
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Disk fails, the agent makes no disks yet.
func (a *Agent) Disk(context.Context, protocol.Disk, io.Writer) error {
	return errors.New("the agent makes no disks yet")
}
