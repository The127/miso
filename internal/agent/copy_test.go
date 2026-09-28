//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// sent are entries as the host sends them, each with its content.
type sent struct {
	entries  []protocol.Entry
	contents []string
}

func (s *sent) Next() (protocol.Entry, io.Reader, error) {
	if len(s.entries) == 0 {
		return protocol.Entry{}, nil, io.EOF
	}

	entry, content := s.entries[0], s.contents[0]
	s.entries, s.contents = s.entries[1:], s.contents[1:]

	return entry, strings.NewReader(content), nil
}

// bareLayers are layers with one empty layer, bare, to build on.
func bareLayers(t *testing.T) string {
	t.Helper()

	layers := t.TempDir()
	work, err := layer.Open(layers).Begin("bare")
	require.NoError(t, err)
	require.NoError(t, work.Finish())

	return layers
}

func TestWhatACopySendsIsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd"}, Destination: "/etc/motd"}
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "copy", "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}

func TestACopiedTreeKeepsItsDirectoriesAndLinks(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"etc"}, Destination: "/etc"}
	files := &sent{
		entries: []protocol.Entry{
			{Kind: "directory", Path: ".", Mode: 0o755},
			{Kind: "link", Path: "issue", Target: "motd"},
			{Kind: "file", Path: "motd", Mode: 0o644, Size: 6},
		},
		contents: []string{"", "", "hello\n"},
	}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(layers, "copy", "etc"))
	target, err := os.Readlink(filepath.Join(layers, "copy", "etc", "issue"))
	require.NoError(t, err)
	assert.Equal(t, "motd", target)
	assert.FileExists(t, filepath.Join(layers, "copy", "etc", "motd"))
}
