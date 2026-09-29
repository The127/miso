package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// received is what a fetch sent to the host.
type received struct {
	length int64
	disk   []byte
	bytes  int64
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
	r.bytes += int64(len(data))

	return nil
}

// keptDisk is a layers directory holding a disk as the output of the key
// "disk".
func keptDisk(t *testing.T, content []byte) string {
	t.Helper()

	layers := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(layers, "disk"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(layers, "disk", "disk.raw"), content, 0o600))

	return layers
}

func TestAFetchSendsTheLengthOfTheDisk(t *testing.T) {
	// arrange
	worker := agent.New(keptDisk(t, make([]byte, 16)), t.TempDir())
	sent := &received{}

	// act
	err := worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, sent, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, int64(16), sent.length)
}

func TestAFetchSendsTheDataOfTheDisk(t *testing.T) {
	// arrange
	worker := agent.New(keptDisk(t, []byte("hello, disk\n")), t.TempDir())
	sent := &received{}

	// act
	err := worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, sent, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hello, disk\n", string(sent.disk))
}

func TestAFetchSkipsTheHolesOfTheDisk(t *testing.T) {
	// arrange
	layers := keptDisk(t, nil)
	disk, err := os.OpenFile(filepath.Join(layers, "disk", "disk.raw"), os.O_WRONLY, 0)
	require.NoError(t, err)
	require.NoError(t, disk.Truncate(1<<20))
	_, err = disk.WriteAt([]byte("hello"), 1<<19)
	require.NoError(t, err)
	require.NoError(t, disk.Close())
	worker := agent.New(layers, t.TempDir())
	sent := &received{}

	// act
	err = worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, sent, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Less(t, sent.bytes, int64(1<<20))
	assert.Equal(t, "hello", string(sent.disk[1<<19:1<<19+5]))
}

func TestAFetchNeverFollowsTheDiskAsALink(t *testing.T) {
	// arrange
	layers := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(layers, "disk"), 0o700))
	require.NoError(t, os.Symlink(outside, filepath.Join(layers, "disk", "disk.raw")))
	worker := agent.New(layers, t.TempDir())
	sent := &received{}

	// act
	err := worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, sent, io.Discard)

	// assert
	assert.ErrorIs(t, err, unix.ELOOP)
	assert.Empty(t, sent.disk)
}

func TestAFetchRefusesADiskThatIsNoRegularFile(t *testing.T) {
	// arrange
	layers := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(layers, "disk"), 0o700))
	require.NoError(t, unix.Mkfifo(filepath.Join(layers, "disk", "disk.raw"), 0o600))
	worker := agent.New(layers, t.TempDir())
	fetched := make(chan error, 1)

	// act
	go func() {
		fetched <- worker.Fetch(context.Background(), protocol.Fetch{Key: "disk"}, &received{}, io.Discard)
	}()

	// assert
	select {
	case err := <-fetched:
		assert.EqualError(t, err, "the disk of disk is no regular file")
	case <-time.After(time.Second):
		assert.Fail(t, "the fetch waits for a writer")
	}
}

func TestAFetchRefusesAKeyThatIsNoLayerName(t *testing.T) {
	// arrange
	outside := keptDisk(t, []byte("outside"))
	layers := filepath.Join(outside, "layers")
	require.NoError(t, os.Mkdir(layers, 0o700))
	worker := agent.New(layers, t.TempDir())
	sent := &received{}

	// act
	err := worker.Fetch(context.Background(), protocol.Fetch{Key: "../disk"}, sent, io.Discard)

	// assert
	assert.ErrorIs(t, err, layer.ErrBadKey)
	assert.Empty(t, sent.disk)
}
