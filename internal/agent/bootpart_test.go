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
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// withImage adds an image layer, which the key image names, that runs a
// command on the base.
func withImage(t *testing.T, worker *agent.Agent, command string) {
	t.Helper()

	image := protocol.Run{Key: "image", Layers: []string{"base"}, Command: command}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
}

func TestAKernelPartIsTheKernelOfTheImageKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	files := "mkdir -p /usr/lib/modules/99.0 && cd /usr/lib/modules/99.0 && echo the-kernel > vmlinuz && echo the-initrd > initrd"
	withImage(t, worker, files)
	part := protocol.BootPart{Key: "kernel", Layers: []string{"base", "image"}, Part: protocol.PartKernel}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "kernel", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "the-kernel\n", string(written))
}

func TestAnInitrdPartIsTheInitrdOfTheImageKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	files := "mkdir -p /usr/lib/modules/99.0 && cd /usr/lib/modules/99.0 && echo the-kernel > vmlinuz && echo the-initrd > initrd"
	withImage(t, worker, files)
	part := protocol.BootPart{Key: "initrd", Layers: []string{"base", "image"}, Part: protocol.PartInitrd}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "initrd", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "the-initrd\n", string(written))
}

func TestABootPartWhoseLayerIsThereAlreadyIsNotKeptAgain(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	for _, key := range []string{"bare", "kernel"} {
		work, err := store.Begin(key)
		require.NoError(t, err)
		require.NoError(t, work.Finish())
	}

	worker := agent.New(layers, t.TempDir())
	// bare has no kernel, so keeping the kernel of it would fail
	part := protocol.BootPart{Key: "kernel", Layers: []string{"bare"}, Part: protocol.PartKernel}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.NoError(t, err)
}

func TestABootPartOfAnImageWithoutAKernelFailsNamingItAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	strip := protocol.Run{Key: "stripped", Layers: []string{"base"}, Command: "rm -r /boot/vmlinuz-* /usr/lib/modules"}
	code, err := worker.Run(context.Background(), strip, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	part := protocol.BootPart{Key: "kernel", Layers: []string{"base", "stripped"}, Part: protocol.PartKernel}

	// act
	err = worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoKernel)
	assert.NoDirExists(t, filepath.Join(layers, "kernel"))
}

func TestAKernelPartIsReadThroughTheImagesOwnLinks(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	files := "mkdir -p /usr/lib/modules/99.0 && cd /usr/lib/modules/99.0 && touch initrd" +
		" && echo the-kernel > /etc/miso-kernel && ln -s /etc/miso-kernel vmlinuz"
	withImage(t, worker, files)
	part := protocol.BootPart{Key: "kernel", Layers: []string{"base", "image"}, Part: protocol.PartKernel}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "kernel", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "the-kernel\n", string(written))
}

func TestABootPartOfAPartTheAgentCannotKeepFailsNamingItAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	part := protocol.BootPart{Key: "part", Layers: []string{"base"}, Part: "banana"}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.ErrorContains(t, err, "banana")
	assert.NoDirExists(t, filepath.Join(layers, "part"))
}
