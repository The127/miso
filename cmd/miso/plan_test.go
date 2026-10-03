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

func TestAPlanTakesTheContextAfterTheDashes(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN echo planned\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "plan", "--", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), "echo planned")
}

func TestAPlanSaysCachedForTheStepsABuildMadeAndRunForTheOnesAfterThem(t *testing.T) {
	// arrange
	dir := t.TempDir()
	// a word of its own, or a cache from an earlier run holds the step
	word := rand.Text()
	file := filepath.Join(dir, "Imagefile")
	require.NoError(t, os.WriteFile(file, []byte("FROM debian:13\nRUN echo "+word+"\n"), 0o600))
	built, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary
	require.NoError(t, err, string(built))
	require.NoError(t, os.WriteFile(file, []byte("FROM debian:13\nRUN echo "+word+"\nRUN echo after "+word+"\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "plan", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Regexp(t, `cached  RUN echo `+word+`\n`, string(said))
	assert.Regexp(t, `run     RUN echo after `+word+`\n`, string(said))
}
