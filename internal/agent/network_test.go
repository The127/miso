//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

// online is the network the VM tests hand a run, as a build hands it the
// network of the builder.
func online(t *testing.T) *protocol.Network {
	t.Helper()

	card := os.Getenv("MISO_VMTEST_MAC")
	require.NotEmpty(t, card, "MISO_VMTEST_MAC names no card")

	return &protocol.Network{Card: card}
}

func TestAnOnlineRunHasACardBesidesItsLoopback(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "ls /sys/class/net"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "eth0\nlo\n", out.String())
}
