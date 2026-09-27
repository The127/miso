package builder_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
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
	err := builder.Ask(running{}, dial, agentName, requests, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, requests, agent.asked)
	assert.Equal(t, 2, *dials)
}

func TestWhatTheAgentWritesReachesTheOutput(t *testing.T) {
	// arrange
	dial, _ := dialling(saying{"unpacking\n"})
	requests := []protocol.Message{
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	}
	var out bytes.Buffer

	// act
	err := builder.Ask(running{}, dial, agentName, requests, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "unpacking\nunpacking\n", out.String())
}

func TestTheAgentIsWaitedForUntilItListens(t *testing.T) {
	// arrange
	agent := &recording{}
	dial, _ := dialling(agent)
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(running{}, booting(2, dial), agentName, requests, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, requests, agent.asked)
}

func TestAVMThatStopsBeforeItsAgentListensFailsWithItsReason(t *testing.T) {
	// arrange
	reason := errors.New("QEMU stopped: exit status 1: could not open kernel")
	dial, _ := dialling(&recording{})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(stopped{reason}, booting(math.MaxInt, dial), agentName, requests, io.Discard)

	// assert
	assert.ErrorIs(t, err, reason)
}

func TestAVMThatStopsWithoutAReasonBeforeItsAgentListensSaysOnlyThat(t *testing.T) {
	// arrange
	dial, _ := dialling(&recording{})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(stopped{}, booting(math.MaxInt, dial), agentName, requests, io.Discard)

	// assert
	assert.EqualError(t, err, "the builder VM stopped before its agent listened")
}

// running is a VM that keeps running.
type running struct{}

func (running) Done() <-chan struct{} { return nil }

func (running) Err() error { return nil }

// stopped is a VM that has stopped for a reason.
type stopped struct {
	reason error
}

func (s stopped) Done() <-chan struct{} {
	done := make(chan struct{})
	close(done)

	return done
}

func (s stopped) Err() error { return s.reason }

// booting fails the first dials, the way dialling a VM fails before its
// agent listens, and dials after that.
func booting(fails int, dial func() (io.ReadWriteCloser, error)) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		if fails > 0 {
			fails--

			return nil, errors.New("connection reset by peer")
		}

		return dial()
	}
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

// saying is an agent whose work succeeds and writes the same words each
// time.
type saying struct {
	words string
}

func (s saying) Run(_ context.Context, _ protocol.Run, out io.Writer) (int, error) {
	_, err := io.WriteString(out, s.words)

	return 0, err
}

func (s saying) Import(_ context.Context, _ protocol.Import, out io.Writer) error {
	_, err := io.WriteString(out, s.words)

	return err
}
