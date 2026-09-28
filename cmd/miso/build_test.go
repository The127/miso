//go:build kvm

package main_test

import (
	"crypto/rand"
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
	// a word of its own, or a cache from an earlier run holds the step and
	// the build prints nothing
	word := rand.Text()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN echo "+word+"\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), word+"\n")
}

func TestAFailingStepOnTheBuilderKernelNamesItsLineOfTheBuildFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	file := filepath.Join(dir, "Imagefile")
	require.NoError(t, os.WriteFile(file, []byte("FROM debian:13\nRUN false\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err)
	assert.Contains(t, string(said), file+":2: RUN false: command failed")
}

func TestACopiedFileOfTheContextIsInTheImageOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	// content of its own, or a cache from an earlier run holds both steps
	// and the build prints nothing
	motd := "hello " + rand.Text() + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "motd"), []byte(motd), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nCOPY motd /etc/motd\nRUN cat /etc/motd\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), motd)
}
