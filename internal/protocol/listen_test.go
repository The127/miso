package protocol_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

var errClosed = errors.New("closed")

type listener struct {
	conns []io.ReadWriteCloser
}

func (l *listener) Accept() (io.ReadWriteCloser, error) {
	if len(l.conns) == 0 {
		return nil, errClosed
	}

	conn := l.conns[0]
	l.conns = l.conns[1:]

	return conn, nil
}

type conn struct {
	io.Reader
	io.Writer
}

func (conn) Close() error { return nil }

func asking(t *testing.T, request protocol.Message) (*protocol.Conn, conn) {
	t.Helper()
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(request))

	return host, conn{&requests, &replies}
}

func TestEachConnectionGetsItsAnswer(t *testing.T) {
	// arrange
	first, firstConn := asking(t, protocol.Run{Command: "true"})
	second, secondConn := asking(t, protocol.Run{Command: "false"})
	listener := &listener{conns: []io.ReadWriteCloser{firstConn, secondConn}}

	// act
	err := protocol.Serve(listener, "miso 1.2.0", runner{})

	// assert
	assert.ErrorIs(t, err, errClosed)
	answer, err := first.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Done{}, answer)
	answer, err = second.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Done{}, answer)
}
