package protocol_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestARunSentIsTheRunReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New(&wire, &wire)

	// act
	sent := conn.Send(protocol.Run{Command: "apt-get update"})
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, protocol.Run{Command: "apt-get update"}, received)
}

func TestAnExitedSentIsTheExitedReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New(&wire, &wire)

	// act
	sent := conn.Send(protocol.Exited{Code: 100})
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, protocol.Exited{Code: 100}, received)
}

func TestAnUnknownMessageIsAnError(t *testing.T) {
	// arrange
	wire := bytes.NewBufferString(`{"Nope":{}}` + "\n")
	conn := protocol.New(wire, wire)

	// act
	_, err := conn.Receive()

	// assert
	assert.ErrorIs(t, err, protocol.ErrUnknownMessage)
}
