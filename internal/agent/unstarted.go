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

// Copy fails naming why the agent did not start.
func (u Unstarted) Copy(context.Context, protocol.Copy, protocol.Entries, io.Writer) error {
	return u.why()
}

// Disk fails naming why the agent did not start.
func (u Unstarted) Disk(context.Context, protocol.Disk, io.Writer) error {
	return u.why()
}

// Rootfs fails naming why the agent did not start.
func (u Unstarted) Rootfs(context.Context, protocol.Rootfs, io.Writer) error {
	return u.why()
}

// BootPart fails naming why the agent did not start.
func (u Unstarted) BootPart(context.Context, protocol.BootPart, io.Writer) error {
	return u.why()
}

// Fetch fails naming why the agent did not start.
func (u Unstarted) Fetch(context.Context, protocol.Fetch, protocol.Pieces, io.Writer) error {
	return u.why()
}

// Prune fails naming why the agent did not start.
func (u Unstarted) Prune(context.Context, protocol.Prune, io.Writer) error {
	return u.why()
}

// Shell fails naming why the agent did not start.
func (u Unstarted) Shell(context.Context, protocol.Shell, io.Reader, io.Writer) (int, error) {
	return 0, u.why()
}

func (u Unstarted) why() error {
	return fmt.Errorf("agent did not start: %w", u.Err)
}
