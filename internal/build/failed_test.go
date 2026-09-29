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

func TestAFailedImportNamesTheLineOfItsFrom(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Line:       3,
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
	}}}
	requests, err := build.Requests(planned, network)
	require.NoError(t, err)
	require.Len(t, requests, 1)

	// act
	failed := requests[0].Failed(protocol.ErrAgentFailed)

	// assert
	assert.EqualError(t, failed, "line 3: FROM debian:13: agent failed")
}

func TestAFailedCheckNamesItsLineAndWhatItPrinted(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS tools\nFROM debian:13\nOUTPUT disk os.raw --tools=tools\nCHECK command -v htop\n")
	requests, err := build.Requests(source, network)
	require.NoError(t, err)
	fetch := requests[len(requests)-1]

	// act
	failed := fetch.CheckFailed(0, 1, "no htop\n")

	// assert
	assert.EqualError(t, failed, "line 4: CHECK command -v htop: exit code 1\nno htop\n")
}
