package agent_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/protocol"
)

func TestARunOnAnAgentThatDidNotStartFailsNamingWhy(t *testing.T) {
	// arrange
	worker := agent.Unstarted{Err: errors.New("mount cache disk miso-cache: no such device")}

	// act
	_, err := worker.Run(context.Background(), protocol.Run{}, io.Discard)

	// assert
	assert.EqualError(t, err, "agent did not start: mount cache disk miso-cache: no such device")
}

func TestAnImportOnAnAgentThatDidNotStartFailsNamingWhy(t *testing.T) {
	// arrange
	worker := agent.Unstarted{Err: errors.New("mount cache disk miso-cache: no such device")}

	// act
	err := worker.Import(context.Background(), protocol.Import{}, io.Discard)

	// assert
	assert.EqualError(t, err, "agent did not start: mount cache disk miso-cache: no such device")
}
