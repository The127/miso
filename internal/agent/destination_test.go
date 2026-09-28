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
	"github.com/The127/miso/internal/protocol"
)

func TestAFileCopiedIntoADirectoryKeepsTheNameOfItsSource(t *testing.T) {
	// arrange
	layers := bareLayers(t)
	worker := agent.New(layers, t.TempDir())
	files := &sent{entries: []protocol.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, contents: []string{"hello\n"}}
	request := protocol.Copy{Key: "copy", Layers: []string{"bare"}, Sources: []string{"etc/motd"}, Digests: []string{files.digest()}, Destination: "/etc/"}

	// act
	err := worker.Copy(context.Background(), request, files, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "copy", "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}
