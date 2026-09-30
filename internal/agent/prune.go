package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// Prune removes the layers the request lets go and says what went.
func (a *Agent) Prune(_ context.Context, request protocol.Prune, out io.Writer) error {
	swept, err := a.layers.Prune(layer.Policy{
		Now:         time.Now(),
		OlderThan:   request.OlderThan,
		KeepStorage: request.KeepStorage,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "removed %d layers, freed %d bytes\n", swept.Count, swept.Bytes)

	return err
}
