//go:build vmtest

package agent_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/copydigest"
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

// digest is what the plan says of these entries, before any is sent.
func (s *sent) digest() string {
	sums := make([]string, 0, len(s.entries))
	for i, entry := range s.entries {
		mode := strconv.FormatUint(uint64(entry.Mode), 8)
		if entry.Kind == "link" {
			mode = ""
		}

		content := sha256.Sum256([]byte(s.contents[i]))
		sums = append(sums, copydigest.Entry{Kind: entry.Kind, Path: entry.Path, Mode: mode, Target: entry.Target}.Sum(hex.EncodeToString(content[:])))
	}

	return copydigest.Of(sums)
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
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd"}, Digests: []string{files.digest()}, Destination: "/etc/motd"}

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
	files := &sent{
		entries: []protocol.Entry{
			{Kind: "directory", Path: ".", Mode: 0o755},
			{Kind: "link", Path: "issue", Target: "motd"},
			{Kind: "file", Path: "motd", Mode: 0o644, Size: 6},
		},
		contents: []string{"", "", "hello\n"},
	}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"etc"}, Digests: []string{files.digest()}, Destination: "/etc"}

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

func TestAnEntryOfAnUnknownKindFailsTheCopyAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"dev"}, Destination: "/dev/null"}
	files := &sent{entries: []protocol.Entry{{Kind: "device", Path: "."}}, contents: []string{""}}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.ErrorContains(t, err, "device")
	assert.NoDirExists(t, filepath.Join(layers, "copy"))
}

func TestACopyWhoseLayerIsThereAlreadyReadsNothing(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	work, err := layer.Open(layers).Begin("copy")
	require.NoError(t, err)
	require.NoError(t, work.Finish())
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd"}, Destination: "/etc/motd"}
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}

	// act
	err = worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Len(t, files.entries, 1)
}

func TestACopyThatIsNotWhatWasPlannedKeepsNoLayer(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	planned := (&sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 8}}, contents: []string{"goodbye\n"}}).digest()
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd"}, Digests: []string{planned}, Destination: "/etc/motd"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.ErrorIs(t, err, agent.ErrNotPlanned)
	assert.NoDirExists(t, filepath.Join(layers, "copy"))
}

func TestAnEntryOfASourceTheCopyDoesNotNameFailsTheCopy(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Source: 1, Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd"}, Digests: []string{files.digest()}, Destination: "/etc/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.ErrorIs(t, err, agent.ErrUnknownSource)
	assert.NoDirExists(t, filepath.Join(layers, "copy"))
}

func TestACopyOntoScratchHasTheModesOfItsSourceAndNoneOfARun(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := agent.New(layers, t.TempDir())
	files := &sent{
		entries:  []protocol.Entry{{Kind: "directory", Path: ".", Mode: 0o755}, {Kind: "directory", Path: "proc", Mode: 0o555}},
		contents: []string{"", ""},
	}
	request := protocol.Copy{Key: "copy", Sources: []string{"rootfs"}, Digests: []string{files.digest()}, Destination: "/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	root, err := os.Lstat(filepath.Join(layers, "copy"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeDir|0o755, root.Mode())
	proc, err := os.Lstat(filepath.Join(layers, "copy", "proc"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeDir|0o555, proc.Mode())
}

func TestAnEntryOfAnEarlierSourceAfterALaterOneFailsTheCopy(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	motd := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	issue := &sent{entries: []protocol.Entry{{Source: 1, Kind: "file", Path: ".", Mode: 0o644, Size: 8}}, contents: []string{"welcome\n"}}
	files := &sent{entries: append(issue.entries, motd.entries...), contents: append(issue.contents, motd.contents...)}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"motd", "issue"}, Digests: []string{motd.digest(), issue.digest()}, Destination: "/etc/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.ErrorIs(t, err, agent.ErrOutOfOrder)
	assert.NoDirExists(t, filepath.Join(layers, "copy"))
}

func TestAnEntryThatLeavesItsSourceFailsTheCopy(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{
		entries:  []protocol.Entry{{Kind: "directory", Path: ".", Mode: 0o755}, {Kind: "file", Path: "../x", Mode: 0o644, Size: 6}},
		contents: []string{"", "hello\n"},
	}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"etc"}, Digests: []string{files.digest()}, Destination: "/etc"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.ErrorIs(t, err, agent.ErrNotLocal)
	assert.NoDirExists(t, filepath.Join(layers, "copy"))
}

func TestALayerBelowThatIsNoPlainNameFailsTheCopy(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"../.."}, Sources: []string{"motd"}, Digests: []string{files.digest()}, Destination: "/motd"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.ErrorIs(t, err, layer.ErrBadKey)
	assert.NoDirExists(t, filepath.Join(layers, "copy"))
}
