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

// unasked are entries a copy from a stage must never ask the host for.
type unasked struct {
	t *testing.T
}

func (u unasked) Next() (protocol.Entry, io.Reader, error) {
	u.t.Fatal("a copy from a stage asked the host for entries")

	return protocol.Entry{}, nil, io.EOF
}

// built is a layer of an earlier stage, built, holding a file at path.
func built(t *testing.T, layers, path, content string) {
	t.Helper()

	work, err := layer.Open(layers).Begin("built")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(work.Dir(), path)), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), path), []byte(content), 0o600))
	require.NoError(t, work.Finish())
}

func TestACopyFromAStageTakesItsFileFromThatStagesLayersAndAsksTheHostForNothing(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	built(t, layers, "etc/motd", "hello\n")
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
