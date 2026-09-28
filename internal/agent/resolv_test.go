//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
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

func TestAnOnlineRunThatWritesInEtcKeepsTheModeAndOwnerOfEtc(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	odd := protocol.Run{Key: "odd", Layers: []string{"base"}, Command: "chmod 0751 /etc && chown 7:8 /etc"}
	code, err := worker.Run(context.Background(), odd, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	run := protocol.Run{Key: "run", Layers: []string{"base", "odd"}, Network: online(t), Command: "touch /etc/x"}

	// act
	code, err = worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code)
	info, err := os.Lstat(filepath.Join(layers, "run", "etc"))
	require.NoError(t, err)
	assert.Equal(t, fs.ModeDir|0o751, info.Mode())
	owner, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, []uint32{7, 8}, []uint32{owner.Uid, owner.Gid})
}
