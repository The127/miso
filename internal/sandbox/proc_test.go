//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestARunSeesItsOwnProcesses(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "cat /proc/1/cmdline"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "/bin/sh\x00-c\x00cat /proc/1/cmdline\x00", out.String())
}

func TestARunLeavesTheKernelSettingsAsItFoundThem(t *testing.T) {
	// arrange
	path := "/proc/sys/vm/swappiness"
	was, err := os.ReadFile(path)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, os.WriteFile(path, was, 0o600)) }) //nolint:gosec // the test names the setting
	require.NoError(t, os.WriteFile(path, []byte("60"), 0o600))
	root := onBase(t)
	run := protocol.Run{Command: "echo 10 > /proc/sys/vm/swappiness"}

	// act
	_, err = sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	setting, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "60\n", string(setting))
}
