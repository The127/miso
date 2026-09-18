package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func TestARunOnScratchBecomesARequestWithItsCommand(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 1)
	assert.Equal(t, "echo hi", requests[0].Command)
}

func TestARequestCarriesTheKeyOfItsStep(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 1)
	assert.Equal(t, "k1", requests[0].Key)
}

func TestARunIsBuiltOnTheBaseLayerOfItsStage(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 1)
	assert.Equal(t, []string{"base"}, requests[0].Layers)
}

func TestARunKeepsTheVariablesSetBeforeItInOrder(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "scratch",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Env{Line: 3, Key: "B", Value: "2"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Run{Line: 4, Command: "echo $A $B"}, Key: "k3", BuiltOn: []string{"k2"}},
		},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 1)
	assert.Equal(t, []string{"A=1", "B=2"}, requests[0].Env)
}

func TestARunIsBuiltOnTheRunsBeforeItLowestFirst(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "scratch",
		Steps: []plan.Step{
			{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "echo bye"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 2)
	assert.Equal(t, []string{"base", "k1"}, requests[1].Layers)
}

func TestARunIsBuiltOnTheCopiesBeforeIt(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "scratch",
		Steps: []plan.Step{
			{Instruction: imagefile.Copy{Line: 2, Sources: []string{"motd"}, Destination: "/etc/motd"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "cat /etc/motd"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 1)
	assert.Equal(t, []string{"base", "k1"}, requests[0].Layers)
}

func TestAnEnvMakesNoLayer(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "scratch",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Run{Line: 3, Command: "echo $A"}, Key: "k2", BuiltOn: []string{"k1"}},
		},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	require.Len(t, requests, 1)
	assert.Equal(t, []string{"base"}, requests[0].Layers)
}
