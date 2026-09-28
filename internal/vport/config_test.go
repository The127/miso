package vport_test

import (
	"io"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vport"
)

// stderr is what the test's process writes to its standard error, until
// the returned func is called.
func stderr(t *testing.T) func() string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	was := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = was })

	return func() string {
		os.Stderr = was
		_ = w.Close()
		said, _ := io.ReadAll(r)

		return string(said)
	}
}

func TestAVMWhosePortCarriesNoSessionSaysNothingOnStandardError(t *testing.T) {
	// arrange
	said := stderr(t)
	host, guest := net.Pipe()
	t.Cleanup(func() { _ = host.Close() })
	_, err := vport.Listen(guest)
	require.NoError(t, err)

	// act
	_, _ = host.Write([]byte("no session here\n"))
	_, _ = io.ReadAll(host)

	// assert
	assert.Empty(t, said())
}

func TestAHostWhosePortCarriesNoSessionSaysNothingOnStandardError(t *testing.T) {
	// arrange
	said := stderr(t)
	host, guest := net.Pipe()
	t.Cleanup(func() { _ = guest.Close() })
	_, err := vport.Connect(host)
	require.NoError(t, err)

	// act
	_, _ = guest.Write([]byte("no session here\n"))
	_, _ = io.ReadAll(guest)

	// assert
	assert.Empty(t, said())
}

func TestASessionOnAPortKeepsNoWatchOnASilentVM(t *testing.T) {
	// arrange
	config := vport.Config()

	// act
	keepsWatch := config.EnableKeepAlive

	// assert
	assert.False(t, keepsWatch)
}
