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

// twoRuns is a stage with a run on line 2 and another on line 3.
func twoRuns() plan.Plan {
	return plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo one"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "echo two"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}
}

func TestAShellBeforeAStepStandsOnTheLayersBelowIt(t *testing.T) {
	// arrange
	planned := twoRuns()

	// act
	requests, err := build.Before(planned, network, 3)

	// assert
	require.NoError(t, err)
	require.NotEmpty(t, requests)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, []string{"base", "k1"}, shell.Layers)
	assert.Len(t, runsOf(requests), 1, "the step itself is not run")
}

func TestALineWithNoStepHasNoShell(t *testing.T) {
	// arrange
	planned := twoRuns()

	// act
	_, err := build.Before(planned, network, 9)

	// assert
	assert.ErrorIs(t, err, build.ErrNoStepAtLine)
	assert.ErrorContains(t, err, "line 9")
}

func TestAShellBeforeAStepHasTheEnvironmentItWouldRunIn(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "MODE", Value: "debug"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "make"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests, err := build.Before(planned, network, 3)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, []string{"MODE=debug"}, shell.Env)
}

func TestAShellBeforeAnOnlineStepHasTheNetworkOfTheHost(t *testing.T) {
	// arrange
	planned := twoRuns()

	// act
	requests, err := build.Before(planned, network, 3)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, &network, shell.Network)
}

func TestAShellBeforeAnOfflineStepHasNoNetwork(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Offline: true, Command: "make test"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Before(planned, network, 2)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Nil(t, shell.Network)
}
