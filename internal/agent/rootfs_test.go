//go:build vmtest

package agent_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// writesRootfs is a mkfs.ext4 that writes rootfs into the last path it is
// given.
const writesRootfs = `for last; do :; done
echo rootfs > "$last"`

// fakeMkfs adds to the base a mkfs.ext4 that runs the script, and answers
// the layers of those tools.
func fakeMkfs(t *testing.T, worker *agent.Agent, script string) []string {
	t.Helper()

	command := fmt.Sprintf("cat > /usr/local/bin/mkfs.ext4 <<'EOF'\n#!/bin/sh\n%s\nEOF\nchmod 755 /usr/local/bin/mkfs.ext4\n", script)
	run := protocol.Run{Key: "mkfs", Layers: []string{"base"}, Command: command}
	code, err := worker.Run(context.Background(), run, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return []string{"base", "mkfs"}
}

func TestARootfsIsWhatItsToolsWriteKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base"}, Tools: fakeMkfs(t, worker, writesRootfs), Format: protocol.FormatExt4}

	// act
	err := worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "rootfs", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "rootfs\n", string(written))
}

// reportsItsInput is a mkfs.ext4 that writes what it was given into the last
// path: the size of that file and its arguments.
const reportsItsInput = `for last; do :; done
size=$(stat -c %s "$last")
echo "$size $*" > "$last"`

// tinyWithManyEntries is a du that sees an image of 10 KiB and 100000
// entries.
const tinyWithManyEntries = `case "$*" in
*--inodes*) printf '100000\t/run/miso/image\n' ;;
*) printf '10\t/run/miso/image\n' ;;
esac`

const entries = 100000

// madeOfTinyImage makes a rootfs of an image the tools see as tiny with many
// entries, and answers what mkfs saw: the size of its file and its inodes.
func madeOfTinyImage(t *testing.T) (size, inodes int) {
	t.Helper()

	layers := t.TempDir()
	worker := mountedBase(t, layers)
	tools := fakeMkfs(t, worker, reportsItsInput)
	command := "cat > /usr/local/bin/du <<'EOF'\n#!/bin/sh\n" + tinyWithManyEntries + "\nEOF\nchmod 755 /usr/local/bin/du\n"
	du := protocol.Run{Key: "du", Layers: tools, Command: command}
	code, err := worker.Run(context.Background(), du, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base"}, Tools: append(tools, "du"), Format: protocol.FormatExt4}
	require.NoError(t, worker.Rootfs(context.Background(), rootfs, io.Discard))
	said, err := os.ReadFile(filepath.Join(layers, "rootfs", "disk.raw"))
	require.NoError(t, err)

	found := regexp.MustCompile(`^(\d+) .* -N (\d+) `).FindStringSubmatch(string(said))
	require.NotNil(t, found, string(said))

	size, _ = strconv.Atoi(found[1])
	inodes, _ = strconv.Atoi(found[2])

	return size, inodes
}

func TestARootfsGetsAnInodeForEveryEntryOfTheImage(t *testing.T) {
	// arrange and act
	_, inodes := madeOfTinyImage(t)

	// assert
	assert.GreaterOrEqual(t, inodes, entries)
}

func TestARootfsFileHasABlockForEveryEntryOfTheImage(t *testing.T) {
	// arrange and act
	size, _ := madeOfTinyImage(t)

	// assert
	assert.GreaterOrEqual(t, size, entries*4096)
}

func TestARootfsIsMadeFromTheImageAsMkfsCopiesIt(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	mark := rand.Text()
	image := protocol.Run{Key: "image", Layers: []string{"base"}, Command: "echo " + mark + " > /etc/miso-image"}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	mkfs := `for arg; do if [ "$prev" = -d ]; then dir=$arg; fi; prev=$arg; last=$arg; done
cat "$dir/etc/miso-image" > "$last"`
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base", "image"}, Tools: fakeMkfs(t, worker, mkfs), Format: protocol.FormatExt4}

	// act
	err = worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "rootfs", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, mark+"\n", string(written))
}

func TestARootfsWhoseLayerIsThereAlreadyIsNotMadeAgain(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	for _, key := range []string{"bare", "rootfs"} {
		work, err := store.Begin(key)
		require.NoError(t, err)
		require.NoError(t, work.Finish())
	}

	worker := agent.New(layers, t.TempDir())
	// bare has no tools, so making a file system of it would fail
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"bare"}, Tools: []string{"bare"}, Format: protocol.FormatExt4}

	// act
	err := worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	assert.NoError(t, err)
}

func TestARootfsWhoseToolsFailFailsNamingTheExitCodeAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base"}, Tools: fakeMkfs(t, worker, "exit 3"), Format: protocol.FormatExt4}

	// act
	err := worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	assert.ErrorContains(t, err, "exit code 3")
	assert.NoDirExists(t, filepath.Join(layers, "rootfs"))
}

// writesErofs is a mkfs.erofs that writes erofs into its destination, the
// first path it is given.
const writesErofs = `for arg; do case $arg in -*) ;; *) dest=${dest:-$arg} ;; esac; done
echo erofs > "$dest"`

func TestAnErofsRootfsIsWhatMkfsErofsWritesKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	tools := fakeMkfs(t, worker, "exit 9")
	command := "cat > /usr/local/bin/mkfs.erofs <<'EOF'\n#!/bin/sh\n" + writesErofs + "\nEOF\nchmod 755 /usr/local/bin/mkfs.erofs\n"
	erofs := protocol.Run{Key: "erofs", Layers: tools, Command: command}
	code, err := worker.Run(context.Background(), erofs, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base"}, Tools: append(tools, "erofs"), Format: protocol.FormatErofs}

	// act
	err = worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "rootfs", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "erofs\n", string(written))
}

func TestARootfsOfAFormatTheAgentCannotMakeFailsNamingItAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base"}, Tools: fakeMkfs(t, worker, writesRootfs), Format: "btrfs"}

	// act
	err := worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	assert.ErrorContains(t, err, "btrfs")
	assert.NoDirExists(t, filepath.Join(layers, "rootfs"))
}

func TestAPortableRootfsIsOneRootPartitionRepartCopiesTheImageInto(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	repart := `for arg; do case $arg in --definitions=*) definitions=${arg#--definitions=} ;; esac; last=$arg; done
cat "$definitions"/*.conf > "$last"`
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base"}, Tools: fakeTools(t, worker, repart), Format: protocol.FormatErofs, Portable: true}

	// act
	err := worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "rootfs", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "[Partition]\nType=root\nFormat=erofs\nCopyFiles=/\nMinimize=guess\n", string(written))
}

func TestAPortableRootfsWithoutAnOsReleaseFailsSayingSoAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	image := protocol.Run{Key: "image", Layers: []string{"base"}, Command: "rm -f /etc/os-release /usr/lib/os-release"}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base", "image"}, Tools: fakeTools(t, worker, writesDisk), Format: protocol.FormatExt4, Portable: true}
	var out strings.Builder

	// act
	err = worker.Rootfs(context.Background(), rootfs, &out)

	// assert
	assert.Error(t, err)
	assert.Contains(t, out.String(), "os-release")
	assert.NoDirExists(t, filepath.Join(layers, "rootfs"))
}

func TestAPortableRootfsWhoseOsReleaseIsALinkToAPlaceOfTheToolsIsMade(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	image := protocol.Run{Key: "image", Layers: []string{"base"}, Command: "rm -f /etc/os-release /usr/lib/os-release && ln -s /not/in/the/tools /etc/os-release"}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	rootfs := protocol.Rootfs{Key: "rootfs", Layers: []string{"base", "image"}, Tools: fakeTools(t, worker, writesDisk), Format: protocol.FormatExt4, Portable: true}

	// act
	err = worker.Rootfs(context.Background(), rootfs, io.Discard)

	// assert
	assert.NoError(t, err)
}
