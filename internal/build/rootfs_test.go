package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func TestARunIsBuiltOnTheBaseLayerOfItsStage(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "debian:13",
		BaseKey: "base",
		Steps:   []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, []string{"base"}, runsOf(requests)[0].Layers)
}

func TestARunIsBuiltOnTheRunsBeforeItLowestFirst(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "debian:13",
		BaseKey: "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "echo bye"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 2)
	assert.Equal(t, []string{"base", "k1"}, runsOf(requests)[1].Layers)
}

func TestARunIsBuiltOnTheCopiesBeforeIt(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "debian:13",
		BaseKey: "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Copy{Line: 2, Sources: []string{"motd"}, Destination: "/etc/motd"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "cat /etc/motd"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, []string{"base", "k1"}, runsOf(requests)[0].Layers)
}

func TestAnEnvMakesNoLayer(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "debian:13",
		BaseKey: "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "echo $A"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, []string{"base"}, runsOf(requests)[0].Layers)
}

func TestAnOutputMakesNoLayer(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "debian:13",
		BaseKey: "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Output{Line: 2, Kind: "disk", Name: "x.raw"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "echo hi"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, []string{"base"}, runsOf(requests)[0].Layers)
}

func TestARunOnScratchIsBuiltOnNoLayer(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "scratch",
		BaseKey: "s",
		Steps:   []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"s"}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Empty(t, runsOf(requests)[0].Layers)
}
