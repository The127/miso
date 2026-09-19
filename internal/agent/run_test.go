//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/basemount"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// mountedBase is an agent keeping its layers in a directory, with the root of
// the VM's base image mounted as the layer base. Mounted and not imported,
// because an import takes seconds and a run needs only a root below it.
func mountedBase(t *testing.T, layers string) *agent.Agent {
	t.Helper()

	require.NotEmpty(t, os.Getenv("MISO_VMTEST_BASE_DIGEST"), "MISO_VMTEST_BASE names no base image")
	base := filepath.Join(layers, "base")
	require.NoError(t, os.Mkdir(base, 0o700))
	require.NoError(t, basemount.Mount("miso-test-base", base))
	// before the layers are removed, which would fail on a read-only mount
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(base, syscall.MNT_DETACH)) })

	return agent.New(layers, t.TempDir())
}

func TestWhatARunWritesIsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	written, err := os.ReadFile(filepath.Join(layers, "run", "x"))
	require.NoError(t, err)
	assert.Equal(t, "hi\n", string(written))
	assert.NoFileExists(t, filepath.Join(layers, "base", "x"))
}

func TestARunOnARootWithoutAShellFailsNamingIt(t *testing.T) {
	// arrange
	layers := t.TempDir()
	work, err := layer.Open(layers).Begin("bare")
	require.NoError(t, err)
	require.NoError(t, work.Finish())
	worker := agent.New(layers, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"bare"}, Command: "true"}

	// act
	_, err = worker.Run(context.Background(), run, io.Discard)

	// assert
	assert.ErrorContains(t, err, "/bin/sh")
}

func TestARunLeavesNoMountPointsInItsLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	work, err := layer.Open(layers).Begin("bare")
	require.NoError(t, err)
	self, err := os.ReadFile("/proc/self/exe")
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(filepath.Join(work.Dir(), "bin"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "bin", "sh"), self, 0o755)) //nolint:gosec // the shell must be executable
	require.NoError(t, work.Finish())
	worker := agent.New(layers, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"bare"}, Command: "true"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code)
	for _, dir := range []string{"proc", "sys", "dev"} {
		assert.NoDirExists(t, filepath.Join(layers, "run", dir))
	}
}

func TestAFailedRunLeavesNoWork(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x; exit 3"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"base"}, names(t, layers))
}

// layerWithF is the layer of a key in a directory, holding only /f with a
// text.
func layerWithF(t *testing.T, layers, key, text string) {
	t.Helper()

	work, err := layer.Open(layers).Begin(key)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "f"), []byte(text), 0o600))
	require.NoError(t, work.Finish())
}

func TestAHigherLayerHidesALowerOne(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	layerWithF(t, layers, "a", "a")
	layerWithF(t, layers, "b", "b")
	run := protocol.Run{Key: "run", Layers: []string{"base", "a", "b"}, Command: "cat /f > /seen"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	seen, err := os.ReadFile(filepath.Join(layers, "run", "seen"))
	require.NoError(t, err)
	assert.Equal(t, "b", string(seen))
}

func TestARunSeesItsLowestLayerUnderManyOthers(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	layerWithF(t, layers, "lowest", "lowest")
	keys := []string{"base", "lowest"}
	// as long as real keys, so their paths fill more than the page overlay
	// options fit in
	for i := range 60 {
		key := fmt.Sprintf("%064d", i)
		work, err := layer.Open(layers).Begin(key)
		require.NoError(t, err)
		require.NoError(t, work.Finish())
		keys = append(keys, key)
	}

	run := protocol.Run{Key: "run", Layers: keys, Command: "cat /f"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "lowest", out.String())
}

func TestTheRootOfARunKeepsTheModeOfTheLayersBelow(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "stat -c %a / > /seen"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	seen, err := os.ReadFile(filepath.Join(layers, "run", "seen"))
	require.NoError(t, err)
	assert.Equal(t, "755\n", string(seen))
}

// ownFileSystem is an empty directory that is a file system of its own.
func ownFileSystem(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, syscall.Mount("tmpfs", dir, "tmpfs", 0, ""))
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, syscall.MNT_DETACH)) })

	return dir
}

func TestARunWorksWithLayersOnAFileSystemOfTheirOwn(t *testing.T) {
	// arrange
	layers := ownFileSystem(t)
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(layers, "run", "x"))
}

func TestARunWorksWithLayersOnASharedFileSystem(t *testing.T) {
	// arrange
	// as systemd leaves every mount, and unlike this VM's
	layers := ownFileSystem(t)
	require.NoError(t, syscall.Mount("", layers, "", syscall.MS_SHARED, ""))
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "echo hi > /x"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(layers, "run", "x"))
}

// overlayDefault sets a default of the overlay module for one test.
func overlayDefault(t *testing.T, parameter, value string) {
	t.Helper()

	path := filepath.Join("/sys/module/overlay/parameters", parameter)
	was, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(value), 0o600))
	t.Cleanup(func() { assert.NoError(t, os.WriteFile(path, was, 0o600)) }) //nolint:gosec // the test names the parameter
}

// names are the names in a directory.
func names(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	found := make([]string, 0, len(entries))
	for _, entry := range entries {
		found = append(found, entry.Name())
	}

	return found
}

func TestADirectoryARunRenamesIsWholeInItsLayer(t *testing.T) {
	// arrange
	overlayDefault(t, "redirect_dir", "Y")
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "mv /etc/apt /etc/moved"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code)
	assert.Equal(t, names(t, filepath.Join(layers, "base", "etc", "apt")), names(t, filepath.Join(layers, "run", "etc", "moved")))
}

func TestAFileARunChangesTheModeOfIsWholeInItsLayer(t *testing.T) {
	// arrange
	overlayDefault(t, "metacopy", "Y")
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "chmod 600 /etc/debian_version"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code)
	was, err := os.ReadFile(filepath.Join(layers, "base", "etc", "debian_version"))
	require.NoError(t, err)
	is, err := os.ReadFile(filepath.Join(layers, "run", "etc", "debian_version"))
	require.NoError(t, err)
	assert.Equal(t, string(was), string(is))
}
