//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/protocol"
)

// importedBase is an agent keeping its layers in a directory, with the base
// image of the VM imported as the layer base.
func importedBase(t *testing.T, layers string) *agent.Agent {
	t.Helper()

	digest := os.Getenv("MISO_VMTEST_BASE_DIGEST")
	require.NotEmpty(t, digest, "MISO_VMTEST_BASE names no base image")
	worker := agent.New(layers, t.TempDir())
	err := worker.Import(context.Background(), protocol.Import{Key: "base", Digest: digest}, io.Discard)
	require.NoError(t, err)

	return worker
}

func TestWhatARunWritesIsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := importedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	written, err := os.ReadFile(filepath.Join(layers, "run", "x"))
	require.NoError(t, err)
	assert.Equal(t, "hi\n", string(written))
	assert.NoFileExists(t, filepath.Join(layers, "base", "x"))
}

func TestAFailedCommandAnswersItsExitCode(t *testing.T) {
	// arrange
	worker := importedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "exit 3"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestAFailedRunLeavesNoWork(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := importedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x; exit 3"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	entries, err := os.ReadDir(layers)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	assert.Equal(t, []string{"base"}, names)
}
