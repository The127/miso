package protocol_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

// disk is a fetched disk held in memory.
type disk struct {
	bytes []byte
}

func (d *disk) WriteAt(p []byte, offset int64) (int, error) {
	end := int(offset) + len(p)
	if end > len(d.bytes) {
		d.bytes = append(d.bytes, make([]byte, end-len(d.bytes))...)
	}

	return copy(d.bytes[offset:], p), nil
}

func (d *disk) Truncate(size int64) error {
	d.bytes = append(d.bytes, make([]byte, int(size)-len(d.bytes))...)

	return nil
}

func TestAFetchedPieceIsWrittenAtItsOffset(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	require.NoError(t, agent.Send(protocol.Piece{Offset: 4, Size: 6}))
	replies.WriteString("hello\n")
	require.NoError(t, agent.Send(protocol.Done{}))
	host := protocol.New("miso 1.2.0", &replies, &requests)
	fetched := &disk{}

	// act
	err := host.AskFetch(protocol.Fetch{Key: "abc"}, fetched, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("\x00\x00\x00\x00hello\n"), fetched.bytes)
}

func TestAFetchedDiskThatEndsInAHoleKeepsItsLength(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	require.NoError(t, agent.Send(protocol.Length{Size: 16}))
	require.NoError(t, agent.Send(protocol.Piece{Offset: 4, Size: 6}))
	replies.WriteString("hello\n")
	require.NoError(t, agent.Send(protocol.Done{}))
	host := protocol.New("miso 1.2.0", &replies, &requests)
	fetched := &disk{}

	// act
	err := host.AskFetch(protocol.Fetch{Key: "abc"}, fetched, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("\x00\x00\x00\x00hello\n\x00\x00\x00\x00\x00\x00"), fetched.bytes)
}
