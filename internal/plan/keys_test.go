package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func steps(t *testing.T, source string) []imagefile.Instruction {
	t.Helper()

	stages := parse(t, source)
	require.Len(t, stages, 1)

	return stages[0].Instructions
}

func TestAChangedRunCommandChangesItsKey(t *testing.T) {
	// arrange
	vim := steps(t, "FROM scratch\nRUN apt-get install vim\n")
	nano := steps(t, "FROM scratch\nRUN apt-get install nano\n")
	require.Len(t, vim, 1)
	require.Len(t, nano, 1)

	// act
	vimKey := plan.StepKey("", vim[0])
	nanoKey := plan.StepKey("", nano[0])

	// assert
	assert.NotEqual(t, vimKey, nanoKey)
}

func TestTheSameRunAfterADifferentStepGetsADifferentKey(t *testing.T) {
	// arrange
	updated := steps(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install vim\n")
	upgraded := steps(t, "FROM scratch\nRUN apt-get upgrade\nRUN apt-get install vim\n")
	require.Len(t, updated, 2)
	require.Len(t, upgraded, 2)

	// act
	updatedKey := plan.StepKey(plan.StepKey("", updated[0]), updated[1])
	upgradedKey := plan.StepKey(plan.StepKey("", upgraded[0]), upgraded[1])

	// assert
	assert.NotEqual(t, updatedKey, upgradedKey)
}

func TestACheckAndARunOfTheSameCommandGetDifferentKeys(t *testing.T) {
	// arrange
	found := steps(t, "FROM scratch\nRUN true\nCHECK true\n")
	require.Len(t, found, 2)
	require.IsType(t, imagefile.Run{}, found[0])
	require.IsType(t, imagefile.Check{}, found[1])

	// act
	runKey := plan.StepKey("", found[0])
	checkKey := plan.StepKey("", found[1])

	// assert
	assert.NotEqual(t, runKey, checkKey)
}
