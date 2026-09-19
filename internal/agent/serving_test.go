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

func TestAnAgentOnADiskThatIsNotThereAnswersWhyItDidNotStart(t *testing.T) {
	// arrange
	dir := t.TempDir()

	// act
	runner := agent.Serving("miso-test-nowhere", dir)

	// assert
	_, err := runner.Run(context.Background(), protocol.Run{}, io.Discard)
	assert.ErrorContains(t, err, "agent did not start")
	assert.ErrorContains(t, err, "miso-test-nowhere")
}

func TestAnAgentOnTheCacheDiskServesAsAStartedOne(t *testing.T) {
	// arrange
	dir := t.TempDir()
	runner := agent.Serving("miso-cache", dir)
	unmountAtEnd(t, dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "layers", "served"), 0o700))

	// act
	err := runner.Import(context.Background(), protocol.Import{Key: "served", Digest: nowhere}, io.Discard)

	// assert
	assert.NoError(t, err)
}
