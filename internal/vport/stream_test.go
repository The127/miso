package vport_test

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vport"
)

func TestWhatTheHostWritesOnAStreamReachesTheGuest(t *testing.T) {
	// arrange
	host, guest := net.Pipe()
	t.Cleanup(func() {
		_ = host.Close()
		_ = guest.Close()
	})

	// act
	dialer, err := vport.Connect(host)
	require.NoError(t, err)
	listener, err := vport.Listen(guest)
	require.NoError(t, err)
	dialed := make(chan error, 1)
	go func() {
		stream, err := dialer.Dial()
		if err == nil {
			_, err = stream.Write([]byte("hello"))
			_ = stream.Close()
		}

		dialed <- err
	}()
	accepted, err := listener.Accept()
	require.NoError(t, err)
	require.NoError(t, <-dialed)
	read := make([]byte, 5)
	_, err = io.ReadFull(accepted, read)
	require.NoError(t, err)
	got := string(read)

	// assert
	assert.Equal(t, "hello", got)
}

func TestDialingAPortWithNoAgentBehindItFails(t *testing.T) {
	// arrange
	host, silent := net.Pipe()
	t.Cleanup(func() {
		_ = host.Close()
		_ = silent.Close()
	})
	go func() { _, _ = io.Copy(io.Discard, silent) }()
	dialer, err := vport.Connect(host)
	require.NoError(t, err)

	// act
	_, err = dialer.Dial()

	// assert
	assert.Error(t, err)
}
