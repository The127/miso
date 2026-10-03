package protocol_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestAFetchSentIsTheFetchReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New("miso 1.2.0", &wire, &wire)
	request := protocol.Fetch{Key: "abc"}

	// act
	sent := conn.Send(request)
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, request, received)
}

func TestAFetchOfANamedFileSentIsTheFetchReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New("miso 1.2.0", &wire, &wire)
	request := protocol.Fetch{Key: "abc", File: "disk.root.raw"}

	// act
	sent := conn.Send(request)
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, request, received)
}
