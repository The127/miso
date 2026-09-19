//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestARunIsCalledLocalhost(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "hostname"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "localhost\n", out.String())
}

func TestARunCannotRenameTheBuilder(t *testing.T) {
	// arrange
	was, err := os.Hostname()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, syscall.Sethostname([]byte(was))) })
	root := onBase(t)
	run := protocol.Run{Command: "hostname renamed"}

	// act
	code, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code)
	name, err := os.Hostname()
	require.NoError(t, err)
	assert.Equal(t, was, name)
}
