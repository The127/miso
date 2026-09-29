package check_test

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

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

// notices hands out the notifications in order, then waits, and remembers
// which machine they were asked from.
type notices struct {
	sent  chan io.ReadWriteCloser
	asked uint32
}

func sending(said ...string) *notices {
	n := &notices{sent: make(chan io.ReadWriteCloser, len(said))}
	for _, text := range said {
		n.sent <- notice{strings.NewReader(text), io.Discard}
	}

	return n
}

func (n *notices) AcceptFrom(cid uint32) (io.ReadWriteCloser, error) {
	n.asked = cid

	return <-n.sent, nil
}

// machine is the context ID of the VM whose notices the tests take.
const machine = 7

// booted waits for Booted, and fails the test at once if it never returns
// rather than when the test run times out.
func booted(t *testing.T, notices check.Notices) error {
	t.Helper()

	done := make(chan error, 1)
	go func() { done <- check.Booted(notices, machine) }()

	select {
	case err := <-done:
		return err
	case <-time.After(time.Second):
		require.FailNow(t, "Booted never returned")

		return nil
	}
}

func TestAnImageHasBootedOnceItsSystemdSaysReady(t *testing.T) {
	// arrange
	said := sending("X_SYSTEMD_MACHINE_ID=52f7605c5e7c40b3a39ffcb0aaceb307", "READY=1\nSTATUS=Ready.")

	// act
	err := booted(t, said)

	// assert
	require.NoError(t, err)
	assert.Empty(t, said.sent)
}

func TestReadyInTheTextOfAStatusIsNoBoot(t *testing.T) {
	// arrange
	said := sending("STATUS=Waiting for READY=1 from the network", "READY=1")

	// act
	err := booted(t, said)

	// assert
	require.NoError(t, err)
	assert.Empty(t, said.sent)
}

// late is a notice sent after the boot, which tells whether it was read to
// its end before it was closed.
type late struct {
	*strings.Reader
	io.Writer

	closed chan bool
}

func (l *late) Close() error {
	l.closed <- l.Len() == 0

	return nil
}

func TestNoticesAfterTheBootAreStillTaken(t *testing.T) {
	// arrange
	said := sending("READY=1")
	require.NoError(t, booted(t, said))
	notice := &late{Reader: strings.NewReader("X_SYSTEMD_UNIT_ACTIVE=getty.target"), Writer: io.Discard, closed: make(chan bool, 1)}

	// act
	said.sent <- notice

	// assert
	select {
	case read := <-notice.closed:
		assert.True(t, read)
	case <-time.After(time.Second):
		assert.Fail(t, "the notice after the boot was never taken")
	}
}

// broken is a notice whose connection fails while it is read.
type broken struct {
	io.Writer
}

func (broken) Read([]byte) (int, error) { return 0, errors.New("connection reset by peer") }

func (broken) Close() error { return nil }

func TestANoticeThatBreaksAfterTheBootDoesNotEndTheTaking(t *testing.T) {
	// arrange
	said := sending("READY=1")
	require.NoError(t, booted(t, said))
	notice := &late{Reader: strings.NewReader("X_SYSTEMD_UNIT_ACTIVE=getty.target"), Writer: io.Discard, closed: make(chan bool, 1)}

	// act
	said.sent <- broken{io.Discard}
	said.sent <- notice

	// assert
	select {
	case read := <-notice.closed:
		assert.True(t, read)
	case <-time.After(time.Second):
		assert.Fail(t, "the notice after the broken one was never taken")
	}
}

// cut is a notice that breaks after it said something.
type cut struct {
	said *strings.Reader
	io.Writer
}

func (c cut) Read(p []byte) (int, error) {
	if c.said.Len() == 0 {
		return 0, errors.New("connection reset by peer")
	}

	return c.said.Read(p)
}

func (cut) Close() error { return nil }

func TestReadyInANoticeThatBreaksAfterwardsStillCounts(t *testing.T) {
	// arrange
	said := &notices{sent: make(chan io.ReadWriteCloser, 1)}
	said.sent <- cut{strings.NewReader("READY=1\nSTATUS=Rea"), io.Discard}

	// act
	err := booted(t, said)

	// assert
	require.NoError(t, err)
}

func TestANoticeThatBreaksBeforeTheBootDoesNotEndTheWait(t *testing.T) {
	// arrange
	said := &notices{sent: make(chan io.ReadWriteCloser, 2)}
	said.sent <- broken{io.Discard}
	said.sent <- notice{strings.NewReader("READY=1"), io.Discard}

	// act
	err := booted(t, said)

	// assert
	require.NoError(t, err)
}

func TestABootTakesTheNoticesOfItsOwnMachineOnly(t *testing.T) {
	// arrange
	said := sending("READY=1")

	// act
	err := booted(t, said)

	// assert
	require.NoError(t, err)
	assert.Equal(t, uint32(machine), said.asked)
}
