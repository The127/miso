package builder_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
)

// file is an output held in memory.
type file struct {
	bytes  []byte
	closed bool
}

func (f *file) WriteAt(p []byte, offset int64) (int, error) {
	if end := int(offset) + len(p); end > len(f.bytes) {
		f.bytes = append(f.bytes, make([]byte, end-len(f.bytes))...)
	}

	return copy(f.bytes[offset:], p), nil
}

func (f *file) Truncate(size int64) error {
	f.bytes = append(f.bytes, make([]byte, int(size)-len(f.bytes))...)

	return nil
}

func (f *file) Close() error {
	f.closed = true

	return nil
}

// sending is an agent that sends a disk for every fetch.
type sending struct {
	recording

	disk string
}

func (s *sending) Fetch(_ context.Context, request protocol.Fetch, pieces protocol.Pieces, _ io.Writer) error {
	s.note(request)

	if err := pieces.Length(int64(len(s.disk))); err != nil {
		return err
	}

	return pieces.Piece(protocol.Piece{Size: int64(len(s.disk))}, strings.NewReader(s.disk))
}

func TestAFetchedDiskIsWrittenIntoTheFileItsOutputNames(t *testing.T) {
	// arrange
	agent := &sending{disk: "disk\n"}
	requests := []build.Request{{Message: protocol.Fetch{Key: "disk"}, Output: "os.raw"}}
	created := map[string]*file{}
	outputs := func(request build.Request) (builder.Output, error) {
		created[request.Output] = &file{}

		return created[request.Output], nil
	}

	// act
	err := builder.Ask(t.Context(), running{}, dialling(agent), agentName, requests, nil, outputs, io.Discard)

	// assert
	require.NoError(t, err)
	require.Contains(t, created, "os.raw")
	assert.Equal(t, "disk\n", string(created["os.raw"].bytes))
	assert.True(t, created["os.raw"].closed)
}

func TestAnOutputInADirectoryIsTheFileOfItsName(t *testing.T) {
	// arrange
	dir := t.TempDir()

	// act
	output, err := builder.OutputsIn(dir)(build.Request{Output: "os.raw"})
	require.NoError(t, err)
	_, err = output.WriteAt([]byte("disk\n"), 0)
	require.NoError(t, err)
	require.NoError(t, output.Close())

	// assert
	written, err := os.ReadFile(filepath.Join(dir, "os.raw"))
	require.NoError(t, err)
	assert.Equal(t, "disk\n", string(written))
}

func TestAnOutputReplacesAnOlderFileOfItsName(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "os.raw"), []byte("XXXXXXXX"), 0o600))

	// act
	output, err := builder.OutputsIn(dir)(build.Request{Output: "os.raw"})
	require.NoError(t, err)
	_, err = output.WriteAt([]byte("ab"), 0)
	require.NoError(t, err)
	require.NoError(t, output.Close())

	// assert
	written, err := os.ReadFile(filepath.Join(dir, "os.raw"))
	require.NoError(t, err)
	assert.Equal(t, "ab", string(written))
}

func TestWithoutOutputsNoDiskIsFetched(t *testing.T) {
	// arrange
	agent := &sending{disk: "disk\n"}
	requests := []build.Request{{Message: protocol.Fetch{Key: "disk"}, Output: "os.raw"}}

	// act
	err := builder.Ask(t.Context(), running{}, dialling(agent), agentName, requests, nil, nil, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Empty(t, agent.asked)
}
