//go:build kvm

package main_test

import (
	"crypto/rand"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAShellBeforeAStepSeesTheStepsBeforeItAndNotTheStep(t *testing.T) {
	// arrange
	dir := t.TempDir()
	word := rand.Text()
	file := "FROM debian:13\nRUN echo " + word + " > /first\nRUN echo second > /second\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(file), 0o600))
	shell := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "3", dir) //nolint:gosec // the test names the binary
	shell.Stdin = strings.NewReader("cat /first\ntest -e /second || echo no\"\"thing\nexit\n")

	// act
	said, err := shell.CombinedOutput()

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), word+"\r\n")
	assert.Contains(t, string(said), "nothing\r\n")
}

func TestTheExitCodeOfAShellIsTheExitCodeOfMiso(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN true\n"), 0o600))
	shell := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "2", dir) //nolint:gosec // the test names the binary
	shell.Stdin = strings.NewReader("exit 3\n")

	// act
	said, err := shell.CombinedOutput()

	// assert
	exited, isExit := errors.AsType[*exec.ExitError](err)
	require.True(t, isExit, string(said))
	assert.Equal(t, 3, exited.ExitCode())
}

func TestAShellBeforeALineWithNoStepFailsBeforeABoot(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN true\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "9", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err)
	assert.Contains(t, string(said), "line 9: no step at this line")
}
