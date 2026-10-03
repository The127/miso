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
		End: "k2",
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

func TestAShellBuildsNoOutputsOfTheSteps(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo one"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Output{Line: 3, Kind: plan.KindRootfs, Name: "os.img"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Check{Line: 4, Command: "true"}, Key: "k3", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Run{Line: 5, Command: "echo two"}, Key: "k4", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests, err := build.Before(planned, network, 5)

	// assert
	require.NoError(t, err)
	for _, request := range requests {
		switch request.Message.(type) {
		case protocol.Disk, protocol.Rootfs, protocol.BootPart, protocol.Fetch:
			t.Errorf("a shell builds %T", request.Message)
		}
	}
}

func TestAShellAtTheEndStandsOnAllTheLayers(t *testing.T) {
	// arrange
	planned := twoRuns()

	// act
	requests, err := build.After(planned, network)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, []string{"base", "k1", "k2"}, shell.Layers)
	assert.Len(t, runsOf(requests), 2)
	assert.Equal(t, &network, shell.Network)
}

func TestAShellAtTheEndBuildsNoOutputsOfTheSteps(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo one"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Output{Line: 3, Kind: plan.KindRootfs, Name: "os.img"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Check{Line: 4, Command: "true"}, Key: "k3", BuiltOn: []string{"k1"}},
		},
		End: "k1",
	}}}

	// act
	requests, err := build.After(planned, network)

	// assert
	require.NoError(t, err)
	for _, request := range requests {
		switch request.Message.(type) {
		case protocol.Disk, protocol.Rootfs, protocol.BootPart, protocol.Fetch:
			t.Errorf("a shell builds %T", request.Message)
		}
	}
}

func TestAShellAtTheEndOfManyStagesStandsOnTheLastOne(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{
		{
			Name: "tools", Base: "debian:13", BaseDigest: "sha256:image", BaseKey: "base",
			Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "make"}, Key: "t1", BuiltOn: []string{"base"}}},
			End:   "t1",
		},
		{
			Name: "final", Base: "debian:13", BaseDigest: "sha256:image", BaseKey: "base",
			Steps: []plan.Step{{Instruction: imagefile.Run{Line: 5, Command: "echo final"}, Key: "f1", BuiltOn: []string{"base"}}},
			End:   "f1",
		},
	}}

	// act
	requests, err := build.After(planned, network)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, []string{"base", "f1"}, shell.Layers)
	assert.Len(t, runsOf(requests), 2, "the earlier stage is built too")
}

func TestAShellBeforeACheckStandsOnTheLayersAsTheyAreThere(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo one"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Output{Line: 3, Kind: plan.KindRootfs, Name: "os.img"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Run{Line: 4, Command: "echo two"}, Key: "k3", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Check{Line: 5, Command: "true"}, Key: "k4", BuiltOn: []string{"k2"}},
		},
		End: "k3",
	}}}

	// act
	requests, err := build.Before(planned, network, 5)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, []string{"base", "k1", "k3"}, shell.Layers)
}

func TestAShellBeforeAnOutputStandsOnTheLayersBelowIt(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo one"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Output{Line: 3, Kind: plan.KindRootfs, Name: "os.img"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
		End: "k1",
	}}}

	// act
	requests, err := build.Before(planned, network, 3)

	// assert
	require.NoError(t, err)
	shell, isShell := requests[len(requests)-1].Message.(protocol.Shell)
	require.True(t, isShell)
	assert.Equal(t, []string{"base", "k1"}, shell.Layers)
}

func TestTheLineOfAFromIsNoStepToStandBefore(t *testing.T) {
	// arrange
	planned := twoRuns()
	planned.Stages[0].Line = 1
	requests, err := build.Requests(planned, network)
	require.NoError(t, err)

	// act
	from, step := build.IsFrom(requests, 1), build.IsFrom(requests, 2)

	// assert
	assert.True(t, from)
	assert.False(t, step)
}
