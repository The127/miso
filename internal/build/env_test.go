package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func TestARunKeepsTheVariablesSetBeforeItInOrder(t *testing.T) {
	// arrange
	planned := plan.Plan{Arch: "amd64", Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Env{Line: 3, Key: "B", Value: "2"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Run{Line: 4, Command: "echo $A $B"}, Key: "k3", BuiltOn: []string{"k2"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, []string{"TARGETARCH=amd64", "A=1", "B=2"}, runsOf(requests)[0].Env)
}

func TestAVariableSetTwiceIsSentOnceWithItsLastValue(t *testing.T) {
	// arrange
	planned := plan.Plan{Arch: "amd64", Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Env{Line: 3, Key: "B", Value: "2"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Env{Line: 4, Key: "A", Value: "3"}, Key: "k3", BuiltOn: []string{"k2"}},
			{Instruction: imagefile.Run{Line: 5, Command: "echo $A $B"}, Key: "k4", BuiltOn: []string{"k3"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, []string{"TARGETARCH=amd64", "A=3", "B=2"}, runsOf(requests)[0].Env)
}

func TestARunSeesTheArchitectureItBuildsFor(t *testing.T) {
	// arrange
	planned := plan.Plan{Arch: "arm64", Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "apt-get install linux-image-$TARGETARCH"}, Key: "k1", BuiltOn: []string{"base"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Contains(t, runsOf(requests)[0].Env, "TARGETARCH=arm64")
}
