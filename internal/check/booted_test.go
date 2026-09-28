package check_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
)

// notice is one notification of the image's systemd, a connection of its
// own that closes once it is sent.
type notice struct {
	io.Reader
	io.Writer
}

func (notice) Close() error { return nil }

// notices hands out the notifications in order, then waits.
type notices struct {
	sent chan io.ReadWriteCloser
}

func sending(said ...string) *notices {
	n := &notices{sent: make(chan io.ReadWriteCloser, len(said))}
	for _, text := range said {
		n.sent <- notice{strings.NewReader(text), io.Discard}
	}

	return n
}

func (n *notices) Accept() (io.ReadWriteCloser, error) {
	return <-n.sent, nil
}

func TestAnImageHasBootedOnceItsSystemdSaysReady(t *testing.T) {
	// arrange
	said := sending("X_SYSTEMD_MACHINE_ID=52f7605c5e7c40b3a39ffcb0aaceb307", "READY=1\nSTATUS=Ready.")

	// act
	err := check.Booted(said)

	// assert
	require.NoError(t, err)
	assert.Empty(t, said.sent)
}

func TestReadyInTheTextOfAStatusIsNoBoot(t *testing.T) {
	// arrange
	said := sending("STATUS=Waiting for READY=1 from the network", "READY=1")

	// act
	err := check.Booted(said)

	// assert
	require.NoError(t, err)
	assert.Empty(t, said.sent)
}
