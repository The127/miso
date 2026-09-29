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
}

func (r *received) Length(size int64) error {
	r.length = size

	return nil
}

func (r *received) Piece(protocol.Piece, io.Reader) error {
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
