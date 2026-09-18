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
