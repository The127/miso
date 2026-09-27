//go:build kvm

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestABuildRunsItsStepOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN echo hi\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), "hi\n")
}
