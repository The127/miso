//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// nowhere is the digest of a base disk the VM does not have.
var nowhere = "sha256:" + strings.Repeat("0", 64)

// baseDigest is the digest of the base disk the VM has.
func baseDigest(t *testing.T) string {
	t.Helper()

	digest := os.Getenv("MISO_VMTEST_BASE_DIGEST")
	require.NotEmpty(t, digest, "MISO_VMTEST_BASE names no base image")

	return digest
}

func TestAnImportedBaseIsALayerHoldingItsRoot(t *testing.T) {
	// arrange
	digest := baseDigest(t)
	layers := t.TempDir()
	worker := agent.New(layers, t.TempDir())

	// act
	err := worker.Import(context.Background(), protocol.Import{Key: "abc", Digest: digest}, io.Discard)

	// assert
	require.NoError(t, err)
	layer, err := os.OpenRoot(filepath.Join(layers, "abc"))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, layer.Close()) })
	release, err := layer.ReadFile("etc/os-release")
	require.NoError(t, err)
	assert.Contains(t, string(release), "ID=debian")
}

func TestAFailedImportLeavesNoWork(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := agent.New(layers, t.TempDir())

	// act
	err := worker.Import(context.Background(), protocol.Import{Key: "abc", Digest: nowhere}, io.Discard)

	// assert
	require.Error(t, err)
	entries, err := os.ReadDir(layers)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestABaseWhoseLayerIsThereIsNotImportedAgain(t *testing.T) {
	// arrange
	layers := t.TempDir()
	work, err := layer.Open(layers).Begin("abc")
	require.NoError(t, err)
	require.NoError(t, work.Finish())
	worker := agent.New(layers, t.TempDir())

	// act
	err = worker.Import(context.Background(), protocol.Import{Key: "abc", Digest: nowhere}, io.Discard)

	// assert
	assert.NoError(t, err)
}
