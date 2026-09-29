package vsockns_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsockns"
)

func TestANamespaceForVMsKeepsItsVsockToItselfOrIsRefused(t *testing.T) {
	// act
	namespace, err := vsockns.Open()

	// assert
	if errors.Is(err, vsockns.ErrNotPrivate) {
		return
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })
	mode, err := namespace.Mode()
	require.NoError(t, err)
	assert.Equal(t, "local", mode)
}

func TestANamespaceThatCannotBeMadeLocalIsRefused(t *testing.T) {
	// arrange
	vsockns.NoChildMode(t)

	// act
	_, err := vsockns.Open()

	// assert
	assert.ErrorIs(t, err, vsockns.ErrNotPrivate)
}

func TestANamespaceWhoseModeDidNotTakeIsRefused(t *testing.T) {
	// arrange
	vsockns.UnheededChildMode(t)

	// act
	_, err := vsockns.Open()

	// assert
	assert.ErrorIs(t, err, vsockns.ErrNotPrivate)
}

// opening is what Open gave, or a failure of the test when it never
// returned.
func opening(t *testing.T) error {
	t.Helper()

	opened := make(chan error, 1)
	go func() {
		namespace, err := vsockns.Open()
		if err == nil {
			_ = namespace.Close()
		}

		opened <- err
	}()

	select {
	case err := <-opened:
		return err
	case <-time.After(5 * time.Second):
		require.FailNow(t, "Open never gave up on its helper")

		return nil
	}
}

func TestAHelperThatDoesNotAnswerIsGivenUpOn(t *testing.T) {
	// arrange
	vsockns.Impatient(t)
	vsockns.HangingChildMode(t)

	// act
	err := opening(t)

	// assert
	require.Error(t, err)
	assert.NotErrorIs(t, err, vsockns.ErrNotPrivate)
}

func TestAHelperThatDoesNotAnswerIsEnded(t *testing.T) {
	// arrange
	vsockns.Impatient(t)
	fifo := vsockns.HangingChildMode(t)
	require.Error(t, opening(t))

	// act
	reader, err := unix.Open(fifo, unix.O_RDONLY|unix.O_NONBLOCK, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = unix.Close(reader) })
	_, err = unix.Read(reader, make([]byte, 1))

	// assert
	// a helper still stuck opening the FIFO counts as its writer
	assert.NotErrorIs(t, err, unix.EAGAIN)
}
