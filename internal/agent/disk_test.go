//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
)

func TestADiskOfAnImageWithoutAKernelFailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	strip := protocol.Run{Key: "stripped", Layers: []string{"base"}, Command: "rm -r /boot/vmlinuz-* /usr/lib/modules"}
	code, err := worker.Run(context.Background(), strip, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "stripped"}, Tools: []string{"base"}}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoKernel)
}

// fakeTools adds to the base a systemd-repart that writes disk into the
// last path it is given, and answers the layers of those tools.
func fakeTools(t *testing.T, worker *agent.Agent) []string {
	t.Helper()

	repart := `printf '#!/bin/sh\nfor last; do :; done\necho disk > "$last"\n' > /usr/local/bin/systemd-repart && chmod 755 /usr/local/bin/systemd-repart`
	run := protocol.Run{Key: "tools", Layers: []string{"base"}, Command: repart}
	code, err := worker.Run(context.Background(), run, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return []string{"base", "tools"}
}

func TestADiskIsWhatItsToolsWriteKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	disk := protocol.Disk{Key: "disk", Layers: []string{"base"}, Tools: fakeTools(t, worker)}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "disk\n", string(written))
}
