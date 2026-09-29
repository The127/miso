//go:build vmtest

package agent_test

import (
	"context"
	"crypto/rand"
	"fmt"
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

// writesDisk is a systemd-repart that writes disk into the last path it is
// given.
const writesDisk = `for last; do :; done
echo disk > "$last"`

// fakeTools adds to the base a systemd-repart that runs the script, and
// answers the layers of those tools.
func fakeTools(t *testing.T, worker *agent.Agent, repart string) []string {
	t.Helper()

	command := fmt.Sprintf("cat > /usr/local/bin/systemd-repart <<'EOF'\n#!/bin/sh\n%s\nEOF\nchmod 755 /usr/local/bin/systemd-repart", repart)
	run := protocol.Run{Key: "tools", Layers: []string{"base"}, Command: command}
	code, err := worker.Run(context.Background(), run, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return []string{"base", "tools"}
}

func TestADiskIsWhatItsToolsWriteKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	disk := protocol.Disk{Key: "disk", Layers: []string{"base"}, Tools: fakeTools(t, worker, writesDisk)}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "disk\n", string(written))
}

func TestADiskIsMadeFromTheImageAsRepartsRoot(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	mark := rand.Text()
	image := protocol.Run{Key: "image", Layers: []string{"base"}, Command: "echo " + mark + " > /etc/miso-image"}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	repart := `for arg; do case $arg in --root=*) root=${arg#--root=} ;; esac; last=$arg; done
cat "$root/etc/miso-image" > "$last"`
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "image"}, Tools: fakeTools(t, worker, repart)}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, mark+"\n", string(written))
}

func TestADiskWhoseToolsFailFailsNamingTheExitCodeAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	disk := protocol.Disk{Key: "disk", Layers: []string{"base"}, Tools: fakeTools(t, worker, "exit 3")}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	assert.ErrorContains(t, err, "exit code 3")
	assert.NoDirExists(t, filepath.Join(layers, "disk"))
}
