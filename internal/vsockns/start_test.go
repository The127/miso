package vsockns_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vsockns"
)

func TestAProgramStartedInTheNamespaceRunsAlongsideMisoAndSaysHowItEnded(t *testing.T) {
	// arrange
	namespace := opened(t)
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = null.Close() })
	// the program ends only once miso got control back and made this
	gate := filepath.Join(t.TempDir(), "go")

	// act
	program, err := namespace.Start([]string{"sh", "-c", `while [ ! -e "$0" ]; do sleep 0.01; done; exit 3`, gate}, null)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(gate, nil, 0o600))
	code, err := program.Wait()

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestAProgramStartedInTheNamespaceCanBeStopped(t *testing.T) {
	// arrange
	namespace := opened(t)
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = null.Close() })
	program, err := namespace.Start([]string{"sleep", "60"}, null)
	require.NoError(t, err)

	// act
	err = program.Kill()

	// assert
	require.NoError(t, err)
	_, err = program.Wait()
	assert.ErrorContains(t, err, "killed")
}

func TestStoppingAProgramThatEndedStopsNothingElse(t *testing.T) {
	// arrange
	namespace := opened(t)
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = null.Close() })
	ended, err := namespace.Start([]string{"true"}, null)
	require.NoError(t, err)
	_, err = ended.Wait()
	require.NoError(t, err)
	// the next program's report may well take the number the ended one had
	gate := filepath.Join(t.TempDir(), "go")
	running, err := namespace.Start([]string{"sh", "-c", `while [ ! -e "$0" ]; do sleep 0.01; done; exit 3`, gate}, null)
	require.NoError(t, err)

	// act
	_ = ended.Kill()

	// assert
	require.NoError(t, os.WriteFile(gate, nil, 0o600))
	code, err := running.Wait()
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestAProgramStartedInTheNamespaceEndsWithIt(t *testing.T) {
	// arrange
	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	read, written, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = read.Close() })
	_, err = namespace.Start([]string{"sleep", "60"}, written)
	require.NoError(t, err)
	// the program holds the pipe open for as long as it lives
	require.NoError(t, written.Close())

	// act
	require.NoError(t, namespace.Close())

	// assert
	_, err = io.ReadAll(read)
	assert.NoError(t, err)
}
