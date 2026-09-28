//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestAnOnlineRunResolvesANameThroughItsResolvConf(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "getent ahostsv4 miso.test"}
	var out bytes.Buffer

	// act
	code, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code, out.String())
	assert.Contains(t, out.String(), "192.0.2.53")
}
