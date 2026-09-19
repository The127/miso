//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

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
