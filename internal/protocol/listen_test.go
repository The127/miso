package protocol_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

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

type blocking struct {
	release chan struct{}
	second  chan struct{}
}

func (b blocking) Run(_ context.Context, run protocol.Run, _ io.Writer) (int, error) {
	if run.Command == "wait" {
		<-b.release

		return 0, nil
	}

	close(b.second)

	return 0, nil
}

func (blocking) Import(context.Context, protocol.Import, io.Writer) error { return nil }

func TestASecondConnectionIsAnsweredWhileTheFirstStillRuns(t *testing.T) {
	// arrange
	_, firstConn := asking(t, protocol.Run{Command: "wait"})
	_, secondConn := asking(t, protocol.Run{Command: "true"})
	listener := &listener{conns: []io.ReadWriteCloser{firstConn, secondConn}}
	runner := blocking{release: make(chan struct{}), second: make(chan struct{})}
	served := make(chan error)

	// act
	go func() { served <- protocol.Serve(listener, "miso 1.2.0", runner) }()

	// assert
	select {
	case <-runner.second:
	case <-time.After(5 * time.Second):
		assert.Fail(t, "the second connection waits for the first")
	}

	close(runner.release)
	assert.ErrorIs(t, <-served, errClosed)
}
