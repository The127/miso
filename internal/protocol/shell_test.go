package protocol_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestAShellGetsWhatTheHostTypes(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Shell{}))
	require.NoError(t, host.Send(protocol.Input{Bytes: []byte("ls\n")}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{})

	// assert
	require.NoError(t, err)
	output, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Output{Bytes: []byte("ls\n")}, output)
}

func TestAShellThatExitsNonZeroIsExited(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Shell{}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{code: 3})

	// assert
	require.NoError(t, err)
	exited, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Exited{Code: 3}, exited)
}

func TestClosingTheConnectionCancelsTheShell(t *testing.T) {
	// arrange
	requests, hostEnd := io.Pipe()
	var replies bytes.Buffer
	host := protocol.New("miso 1.2.0", bytes.NewReader(nil), hostEnd)
	agent := protocol.New("miso 1.2.0", requests, &replies)
	runner := &waiting{}
	go func() {
		_ = host.Send(protocol.Shell{})
		_ = hostEnd.Close()
	}()

	// act
	_ = agent.Serve(runner)

	// assert
	assert.True(t, runner.cancelled)
}

func TestAnythingButInputCancelsTheShell(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Shell{}))
	require.NoError(t, host.Send(protocol.Run{Command: "true"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	runner := &waiting{}

	// act
	_ = agent.Serve(runner)

	// assert
	assert.True(t, runner.cancelled)
}

// sizing is an agent whose shell notes the first size the terminal changes
// to.
type sizing struct {
	runner
	got *protocol.Resize
}

func (s sizing) Shell(_ context.Context, _ protocol.Shell, term protocol.Terminal, _ io.Writer) (int, error) {
	select {
	case size := <-term.Resized:
		*s.got = size
	case <-time.After(time.Second):
	}

	return 0, nil
}

func TestAShellGetsTheSizesTheTerminalOfTheHostChangesTo(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Shell{}))
	require.NoError(t, host.Send(protocol.Resize{Rows: 30, Cols: 100}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	var got protocol.Resize

	// act
	err := agent.Serve(sizing{got: &got})

	// assert
	require.NoError(t, err)
	assert.Equal(t, protocol.Resize{Rows: 30, Cols: 100}, got)
}
