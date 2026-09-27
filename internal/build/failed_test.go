package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

func TestAFailedRunNamesItsLineAndHowItWasWritten(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "false"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}
	requests, err := build.Requests(planned, network)
	require.NoError(t, err)
	require.Len(t, requests, 2)
	run := requests[1]

	// act
	failed := run.Failed(protocol.ErrCommandFailed)

	// assert
	assert.EqualError(t, failed, "line 2: RUN false: command failed")
	assert.ErrorIs(t, failed, protocol.ErrCommandFailed)
}
