//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestARunSeesTheSystem(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "ls /sys/class/net"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "lo\n")
}

func TestARunLeavesTheSystemAsItFoundIt(t *testing.T) {
	// arrange
	overlayDefault(t, "redirect_dir", "N")
	root := onBase(t)
	run := protocol.Run{Command: "echo Y > /sys/module/overlay/parameters/redirect_dir"}

	// act
	_, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	setting, err := os.ReadFile("/sys/module/overlay/parameters/redirect_dir")
	require.NoError(t, err)
	assert.Equal(t, "N\n", string(setting))
}

// overlayDefault sets a default of the overlay module for one test.
func overlayDefault(t *testing.T, parameter, value string) {
	t.Helper()

	path := filepath.Join("/sys/module/overlay/parameters", parameter)
	was, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(value), 0o600))
	t.Cleanup(func() { assert.NoError(t, os.WriteFile(path, was, 0o600)) }) //nolint:gosec // the test names the parameter
}
