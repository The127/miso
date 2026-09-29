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
	"github.com/The127/miso/internal/protocol"
)

// received is what a fetch sent to the host.
type received struct {
	length int64
	disk   []byte
}

func (r *received) Length(size int64) error {
	r.length = size

	return nil
}

func (r *received) Piece(piece protocol.Piece, content io.Reader) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	if end := int(piece.Offset) + len(data); end > len(r.disk) {
		r.disk = append(r.disk, make([]byte, end-len(r.disk))...)
	}

	copy(r.disk[piece.Offset:], data)

	return nil
}

// keptDisk is a layers directory holding a disk as the output of the key.
func keptDisk(t *testing.T, key string, content []byte) string {
	t.Helper()

	layers := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(layers, key), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(layers, key, "disk.raw"), content, 0o600))

	return layers
}

func TestAFetchSendsTheLengthOfTheDisk(t *testing.T) {
	// arrange
	worker := agent.New(keptDisk(t, "disk", make([]byte, 16)), t.TempDir())
	sent := &received{}

	// act
	err := worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, sent, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, int64(16), sent.length)
}

func TestAFetchSendsTheDataOfTheDisk(t *testing.T) {
	// arrange
	worker := agent.New(keptDisk(t, "disk", []byte("hello, disk\n")), t.TempDir())
	sent := &received{}

	// act
	err := worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, sent, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hello, disk\n", string(sent.disk))
}
