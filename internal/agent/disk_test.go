package agent_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/protocol"
)

func TestADiskFailsBecauseTheAgentMakesNoDisksYet(t *testing.T) {
	// arrange
	worker := agent.New(t.TempDir(), t.TempDir())

	// act
	err := worker.Disk(context.Background(), protocol.Disk{}, io.Discard)

	// assert
	assert.EqualError(t, err, "the agent makes no disks yet")
}
