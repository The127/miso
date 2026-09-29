package protocol_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestTheBytesOfAPieceFollowIt(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	agent := protocol.New("miso 1.2.0", &wire, &wire)
	host := protocol.New("miso 1.2.0", &wire, &wire)
	piece := protocol.Piece{Offset: 4096, Size: 6}
	require.NoError(t, agent.Send(piece))
	wire.WriteString("hello\n")
	require.NoError(t, agent.Send(protocol.Done{}))

	// act
	received, err := host.Receive()
	require.NoError(t, err)
	content, err := io.ReadAll(host.Content())
	require.NoError(t, err)
	next, err := host.Receive()

	// assert
	require.NoError(t, err)
	assert.Equal(t, piece, received)
	assert.Equal(t, "hello\n", string(content))
	assert.Equal(t, protocol.Done{}, next)
}

func TestAPieceWithANegativeSizeIsRefused(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	agent := protocol.New("miso 1.2.0", &wire, &wire)
	host := protocol.New("miso 1.2.0", &wire, &wire)
	require.NoError(t, agent.Send(protocol.Piece{Offset: 4096, Size: -1}))

	// act
	_, err := host.Receive()

	// assert
	assert.ErrorIs(t, err, protocol.ErrNegativeSize)
}
