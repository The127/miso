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

func TestACopyAskedForItsEntriesSendsThemAndThenSent(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	require.NoError(t, agent.Send(protocol.Send{}))
	require.NoError(t, agent.Send(protocol.Done{}))
	host := protocol.New("miso 1.2.0", &replies, &requests)
	request := protocol.Copy{Key: "abc", Sources: []string{"motd"}, Destination: "/etc/motd"}
	entry := protocol.Entry{Kind: "file", Path: ".", Mode: 0o644, Size: 6}
	files := func(send func(protocol.Entry, io.Reader) error) error {
		return send(entry, strings.NewReader("hello\n"))
	}

	// act
	err := host.AskCopy(request, files, io.Discard)

	// assert
	require.NoError(t, err)
	asked, err := agent.Receive()
	require.NoError(t, err)
	received, err := agent.Receive()
	require.NoError(t, err)
	content, err := io.ReadAll(agent.Content())
	require.NoError(t, err)
	sent, err := agent.Receive()
	require.NoError(t, err)
	assert.Equal(t, request, asked)
	assert.Equal(t, entry, received)
	assert.Equal(t, "hello\n", string(content))
	assert.Equal(t, protocol.Sent{}, sent)
}
