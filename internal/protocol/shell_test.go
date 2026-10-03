package protocol_test

import (
	"bytes"
	"io"
	"testing"

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
