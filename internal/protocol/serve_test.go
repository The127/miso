package protocol_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

type runner struct {
	writes string
	code   int
}

func (r runner) Run(_ protocol.Run, out io.Writer) (int, error) {
	if r.writes == "" {
		return r.code, nil
	}

	_, err := io.WriteString(out, r.writes)

	return r.code, err
}

func TestARunThatWorksIsDone(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "echo hello"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{writes: "hello\n"})

	// assert
	require.NoError(t, err)
	output, err := host.Receive()
	require.NoError(t, err)
	done, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Output{Bytes: []byte("hello\n")}, output)
	assert.Equal(t, protocol.Done{}, done)
}

func TestARunThatExitsNonZeroIsExited(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "apt-get install nope"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{code: 100})

	// assert
	require.NoError(t, err)
	exited, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Exited{Code: 100}, exited)
}
