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
	conn := protocol.New("miso 1.2.0", &wire, &wire)

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
	conn := protocol.New("miso 1.2.0", &wire, &wire)

	// act
	sent := conn.Send(protocol.Exited{Code: 100})
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, protocol.Exited{Code: 100}, received)
}

func TestAnOutputSentIsTheOutputReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New("miso 1.2.0", &wire, &wire)

	// act
	sent := conn.Send(protocol.Output{Bytes: []byte("50%\r100%\n")})
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, protocol.Output{Bytes: []byte("50%\r100%\n")}, received)
}

func TestAFailedSentIsTheFailedReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New("miso 1.2.0", &wire, &wire)

	// act
	sent := conn.Send(protocol.Failed{Reason: "mount overlay: no space left on device"})
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, protocol.Failed{Reason: "mount overlay: no space left on device"}, received)
}

func TestADoneSentIsTheDoneReceived(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	conn := protocol.New("miso 1.2.0", &wire, &wire)

	// act
	sent := conn.Send(protocol.Done{})
	received, err := conn.Receive()

	// assert
	require.NoError(t, sent)
	require.NoError(t, err)
	assert.Equal(t, protocol.Done{}, received)
}

func TestAnUnknownMessageIsAnError(t *testing.T) {
	// arrange
	wire := bytes.NewBufferString(`{"Agent":"miso 1.2.0","Nope":{}}` + "\n")
	conn := protocol.New("miso 1.2.0", wire, wire)

	// act
	_, err := conn.Receive()

	// assert
	assert.ErrorIs(t, err, protocol.ErrUnknownMessage)
}

func TestAMessageFromAnotherAgentIsRefused(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	sender := protocol.New("miso 1.2.0", &wire, &wire)
	receiver := protocol.New("miso 1.3.0", &wire, &wire)
	require.NoError(t, sender.Send(protocol.Exited{Code: 0}))

	// act
	_, err := receiver.Receive()

	// assert
	assert.ErrorIs(t, err, protocol.ErrAnotherAgent)
	assert.ErrorContains(t, err, "miso 1.2.0")
	assert.ErrorContains(t, err, "miso 1.3.0")
}
