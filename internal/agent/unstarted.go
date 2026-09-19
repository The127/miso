package agent

import (
	"context"
	"fmt"
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Unstarted answers the host in place of an agent that did not start, so
// the host learns why.
type Unstarted struct {
	Err error
}

// Run fails naming why the agent did not start.
func (u Unstarted) Run(context.Context, protocol.Run, io.Writer) (int, error) {
	return 0, u.why()
}

// Import fails naming why the agent did not start.
func (u Unstarted) Import(context.Context, protocol.Import, io.Writer) error {
	return u.why()
}

func (u Unstarted) why() error {
	return fmt.Errorf("agent did not start: %w", u.Err)
}
