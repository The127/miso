//go:build kvm

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPruneAsksTheBuilderAndSaysWhatItRemoved(t *testing.T) {
	// arrange
	// a limit no layer of the cache is old enough to pass, ten years

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "prune", "--older-than", "87600h").CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), "removed 0 layers")
}

func TestAPruneRemovesTheDownloadsNoBuildUsed(t *testing.T) {
	// arrange
	cache := t.TempDir()
	blobs := filepath.Join(cache, "miso", "bases", "sha256")
	require.NoError(t, os.MkdirAll(blobs, 0o750))

	old := filepath.Join(blobs, strings.Repeat("ab", 32))
	require.NoError(t, os.WriteFile(old, []byte("old"), 0o600))

	longAgo := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(old, longAgo, longAgo))

	prune := exec.CommandContext(t.Context(), miso(t), "prune", "--older-than", "24h") //nolint:gosec // the test names the binary
	prune.Env = append(os.Environ(), "XDG_CACHE_HOME="+cache)

	// act
	said, err := prune.CombinedOutput()

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), "removed 1 downloads, freed 3 bytes")
	assert.NoFileExists(t, old)
}
