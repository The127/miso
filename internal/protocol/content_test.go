package protocol_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestAFilesContentFollowsItsEntry(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	host := protocol.New("miso 1.2.0", &wire, &wire)
	agent := protocol.New("miso 1.2.0", &wire, &wire)
	entry := protocol.Entry{Kind: "file", Path: "motd", Mode: 0o644, Size: 6}

	// act
	require.NoError(t, host.SendEntry(entry, strings.NewReader("hello\n")))
	require.NoError(t, host.Send(protocol.Done{}))
	received, err := agent.Receive()
	require.NoError(t, err)
	content, err := io.ReadAll(agent.Content())
	require.NoError(t, err)
	next, err := agent.Receive()

	// assert
	require.NoError(t, err)
	assert.Equal(t, entry, received)
	assert.Equal(t, "hello\n", string(content))
	assert.Equal(t, protocol.Done{}, next)
}

func TestContentLeftUnreadIsSkippedByTheNextReceive(t *testing.T) {
	// arrange
	var wire bytes.Buffer
	host := protocol.New("miso 1.2.0", &wire, &wire)
	agent := protocol.New("miso 1.2.0", &wire, &wire)
	require.NoError(t, host.SendEntry(protocol.Entry{Kind: "file", Path: "motd", Size: 6}, strings.NewReader("hello\n")))
	require.NoError(t, host.Send(protocol.Done{}))
	_, err := agent.Receive()
	require.NoError(t, err)

	// act
	next, err := agent.Receive()

	// assert
	require.NoError(t, err)
	assert.Equal(t, protocol.Done{}, next)
}
