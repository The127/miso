//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// unasked are entries a copy from a stage must never ask the host for.
type unasked struct {
	t *testing.T
}

func (u unasked) Next() (protocol.Entry, io.Reader, error) {
	u.t.Fatal("a copy from a stage asked the host for entries")

	return protocol.Entry{}, nil, io.EOF
}

// built is a layer of an earlier stage, built, holding files at their
// paths.
func built(t *testing.T, layers string, files map[string]string) {
	t.Helper()

	work, err := layer.Open(layers).Begin("built")
	require.NoError(t, err)

	for path, content := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(work.Dir(), path)), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), path), []byte(content), 0o600))
	}

	require.NoError(t, work.Finish())
}

func TestACopyFromAStageTakesItsFileFromThatStagesLayersAndAsksTheHostForNothing(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	built(t, layers, map[string]string{"etc/motd": "hello\n"})
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Stage: "build", From: []string{"built"}, Sources: []string{"/etc/motd"}, Destination: "/motd"}

	// act
	err := worker.Copy(context.Background(), request, unasked{t}, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "copy", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}

func TestACopyFromAStageLeavesOutWhatAHigherLayerOfThatStageRemoved(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	built(t, layers, map[string]string{"etc/motd": "hello\n", "etc/issue": "welcome\n"})
	removing, err := layer.Open(layers).Begin("removing")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(removing.Dir(), "etc"), 0o750))
	// how overlay writes a removal into a layer
	require.NoError(t, unix.Mknod(filepath.Join(removing.Dir(), "etc", "motd"), unix.S_IFCHR, 0))
	require.NoError(t, removing.Finish())
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Stage: "build", From: []string{"built", "removing"}, Sources: []string{"/etc"}, Destination: "/etc"}

	// act
	err = worker.Copy(context.Background(), request, unasked{t}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(layers, "copy", "etc", "issue"))
	assert.NoFileExists(t, filepath.Join(layers, "copy", "etc", "motd"))
}

func TestACopyFromAStageThatFailsKeepsNoLayer(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	built(t, layers, map[string]string{"etc/motd": "hello\n"})
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Stage: "build", From: []string{"built"}, Sources: []string{"/nope"}, Destination: "/nope"}

	// act
	err := worker.Copy(context.Background(), request, unasked{t}, io.Discard)

	// assert
	require.ErrorIs(t, err, fs.ErrNotExist)
	kept, err := layer.Open(layers).Has("copy")
	require.NoError(t, err)
	assert.False(t, kept)
}
