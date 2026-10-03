package agent

import (
	"context"
	"fmt"
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Cached says which of the keys have a finished layer. It marks none of
// them as used, since a plan only looks.
func (a *Agent) Cached(_ context.Context, request protocol.Cached, out io.Writer) error {
	for _, key := range request.Keys {
		there, err := a.layers.Has(key)
		if err != nil {
			return err
		}

		if there {
			if _, err := fmt.Fprintln(out, key); err != nil {
				return err
			}
		}
	}

	return nil
}
