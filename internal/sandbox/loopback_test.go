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

func TestARunCannotChangeTheBuildersNetwork(t *testing.T) {
	// arrange
	flags := "/sys/class/net/lo/flags"
	was, err := os.ReadFile(flags)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, os.WriteFile(flags, was, 0o600)) }) //nolint:gosec // the test names the setting
	root := onBase(t)
	flip := "if ip link show lo | grep -q ,UP; then ip link set lo down; else ip link set lo up; fi"
	run := protocol.Run{Command: flip}

	// act
	code, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code)
	is, err := os.ReadFile(flags)
	require.NoError(t, err)
	assert.Equal(t, string(was), string(is))
}

func TestARunHasALoopback(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "cat /sys/class/net/lo/flags"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	// up and loopback
	assert.Equal(t, "0x9\n", out.String())
}
