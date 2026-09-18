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
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "debian:13",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Env{Line: 3, Key: "B", Value: "2"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Run{Line: 4, Command: "echo $A $B"}, Key: "k3", BuiltOn: []string{"k2"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, []string{"A=1", "B=2"}, requests[0].Env)
}

func TestAVariableSetTwiceIsSentOnceWithItsLastValue(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "debian:13",
		Steps: []plan.Step{
			{Instruction: imagefile.Env{Line: 2, Key: "A", Value: "1"}, Key: "k1", BuiltOn: []string{"base"}},
			{Instruction: imagefile.Env{Line: 3, Key: "B", Value: "2"}, Key: "k2", BuiltOn: []string{"k1"}},
			{Instruction: imagefile.Env{Line: 4, Key: "A", Value: "3"}, Key: "k3", BuiltOn: []string{"k2"}},
			{Instruction: imagefile.Run{Line: 5, Command: "echo $A $B"}, Key: "k4", BuiltOn: []string{"k3"}},
		},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, []string{"A=3", "B=2"}, requests[0].Env)
}
