//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
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

// layerWithF is the layer of a key in a directory, holding only /f with a
// text.
func layerWithF(t *testing.T, layers, key, text string) {
	t.Helper()

	work, err := layer.Open(layers).Begin(key)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "f"), []byte(text), 0o600))
	require.NoError(t, work.Finish())
}

func TestAHigherLayerHidesALowerOne(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := importedBase(t, layers)
	layerWithF(t, layers, "a", "a")
	layerWithF(t, layers, "b", "b")
	run := protocol.Run{Key: "run", Layers: []string{"base", "a", "b"}, Command: "cat /f > /seen"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	seen, err := os.ReadFile(filepath.Join(layers, "run", "seen"))
	require.NoError(t, err)
	assert.Equal(t, "b", string(seen))
}

func TestARunSeesTheEnvironmentOfTheBuildFileAndNotTheAgents(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := importedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Env: []string{"GREETING=hi"}, Command: "env > /seen"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	seen, err := os.ReadFile(filepath.Join(layers, "run", "seen"))
	require.NoError(t, err)
	assert.Contains(t, string(seen), "GREETING=hi\n")
	assert.NotContains(t, string(seen), "MISO_VMTEST_BASE_DIGEST")
}

func TestARunWithoutEnvironmentFindsCommandsOnTheUsualPath(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := importedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "env > /seen"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	seen, err := os.ReadFile(filepath.Join(layers, "run", "seen"))
	require.NoError(t, err)
	assert.Contains(t, string(seen), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\n")
	assert.NotContains(t, string(seen), "MISO_VMTEST_BASE_DIGEST")
}

func TestTheRootOfARunKeepsTheModeOfTheLayersBelow(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := importedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "stat -c %a / > /seen"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	seen, err := os.ReadFile(filepath.Join(layers, "run", "seen"))
	require.NoError(t, err)
	assert.Equal(t, "755\n", string(seen))
}

func TestAProcessARunLeavesBehindDoesNotOutliveIt(t *testing.T) {
	// arrange
	worker := importedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "sleep 1000 &"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

// ownFileSystem is an empty directory that is a file system of its own.
func ownFileSystem(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, syscall.Mount("tmpfs", dir, "tmpfs", 0, ""))
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, syscall.MNT_DETACH)) })

	return dir
}

func TestARunWorksWithLayersOnAFileSystemOfTheirOwn(t *testing.T) {
	// arrange
	layers := ownFileSystem(t)
	worker := importedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(layers, "run", "x"))
}
