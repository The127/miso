package builder_test

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
)

const agentName = "miso 1.2.0"

func TestEveryRequestIsAskedInOrderEachOnItsOwnConnection(t *testing.T) {
	// arrange
	agent := &recording{}
	dial, dials := dialling(agent)
	requests := []protocol.Message{
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	}

	// act
	err := builder.Ask(dial, agentName, requests)

	// assert
	require.NoError(t, err)
	assert.Equal(t, requests, agent.asked)
	assert.Equal(t, 2, *dials)
}

// dialling hands out connections to an agent that does its work with the
// runner, one request per connection as the real one answers, and counts
// them.
func dialling(runner protocol.Runner) (func() (io.ReadWriteCloser, error), *int) {
	dials := 0
	dial := func() (io.ReadWriteCloser, error) {
		dials++
		host, agent := net.Pipe()
		go func() {
			_ = protocol.New(agentName, agent, agent).Serve(runner)
			_ = agent.Close()
		}()

		return host, nil
	}

	return dial, &dials
}

// recording is an agent whose work succeeds and who notes what it was asked.
type recording struct {
	mu    sync.Mutex
	asked []protocol.Message
}

func (r *recording) Run(_ context.Context, run protocol.Run, _ io.Writer) (int, error) {
	r.note(run)

	return 0, nil
}

func (r *recording) Import(_ context.Context, request protocol.Import, _ io.Writer) error {
	r.note(request)

	return nil
}

func (r *recording) note(request protocol.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.asked = append(r.asked, request)
}
