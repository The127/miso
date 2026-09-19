//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
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
