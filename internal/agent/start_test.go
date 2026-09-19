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

func TestAStartedAgentHasSweptWhatAStoppedOneLeft(t *testing.T) {
	// arrange
	dir := onCache(t, func(dir string) {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "layers", "work-left"), 0o700))
	})

	// act
	_, err := agent.Start("miso-cache", dir)
	unmountAtEnd(t, dir)

	// assert
	require.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(dir, "layers", "work-left"))
}

func TestAnAgentStartsOnACacheDiskThatHoldsNoLayersYet(t *testing.T) {
	// arrange
	dir := onCache(t, func(dir string) {
		require.NoError(t, os.RemoveAll(filepath.Join(dir, "layers")))
	})

	// act
	_, err := agent.Start("miso-cache", dir)
	unmountAtEnd(t, dir)

	// assert
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(dir, "layers"))
}

func TestAStartedAgentKeepsAnImportedBaseOnTheCacheDisk(t *testing.T) {
	// arrange
	digest := baseDigest(t)
	dir := t.TempDir()
	worker, err := agent.Start("miso-cache", dir)
	require.NoError(t, err)
	unmountAtEnd(t, dir)
	require.NotNil(t, worker)

	// act
	err = worker.Import(context.Background(), protocol.Import{Key: "started", Digest: digest}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, "layers", "started", "etc", "os-release"))
}
