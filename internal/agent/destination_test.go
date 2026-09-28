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
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

func TestAFileCopiedIntoADirectoryKeepsTheNameOfItsSource(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"etc/motd"}, Digests: []string{files.digest(t)}, Destination: "/etc/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "copy", "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}

func TestEachSourceOfACopyLandsUnderItsOwnName(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	motd := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	issue := &sent{entries: []protocol.Entry{{Source: 1, Kind: "file", Path: ".", Mode: 0o644, Size: 8}}, contents: []string{"welcome\n"}}
	files := &sent{entries: append(motd.entries, issue.entries...), contents: append(motd.contents, issue.contents...)}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd", "issue"}, Digests: []string{motd.digest(t), issue.digest(t)}, Destination: "/etc/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "copy", "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
	written, err = os.ReadFile(filepath.Join(layers, "copy", "etc", "issue"))
	require.NoError(t, err)
	assert.Equal(t, "welcome\n", string(written))
}

func TestAFileCopiedOntoADirectoryOfTheImageGoesIntoIt(t *testing.T) {
	// arrange
	layers := t.TempDir()
	work, err := layer.Open(layers).Begin("etc")
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(filepath.Join(work.Dir(), "etc"), 0o700))
	require.NoError(t, work.Finish())
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"etc"}, Sources: []string{"motd"}, Digests: []string{files.digest(t)}, Destination: "/etc"}

	// act
	err = worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "copy", "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}

func TestASourceIsNamedAsTheHostReadsItsPath(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd/."}, Digests: []string{files.digest(t)}, Destination: "/etc/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(layers, "copy", "etc", "motd"))
}
