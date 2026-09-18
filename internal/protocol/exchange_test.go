package protocol_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestTheOutputOfARunReachesTheWriter(t *testing.T) {
	// arrange
	var replies bytes.Buffer
	agent := protocol.New("miso 1.2.0", bytes.NewReader(nil), &replies)
	require.NoError(t, agent.Send(protocol.Output{Bytes: []byte("hello\n")}))
	require.NoError(t, agent.Send(protocol.Done{}))
	host := protocol.New("miso 1.2.0", &replies, io.Discard)
	var out bytes.Buffer

	// act
	err := host.Run(protocol.Run{Command: "echo hello"}, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hello\n", out.String())
}
