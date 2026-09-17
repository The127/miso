package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func runs(t *testing.T, source string) []imagefile.Run {
	t.Helper()

	stages := parse(t, source)
	require.Len(t, stages, 1)

	var found []imagefile.Run
	for _, instruction := range stages[0].Instructions {
		run, isRun := instruction.(imagefile.Run)
		require.True(t, isRun)

		found = append(found, run)
	}

	return found
}

func TestAChangedRunCommandChangesItsKey(t *testing.T) {
	// arrange
	vim := runs(t, "FROM scratch\nRUN apt-get install vim\n")
	nano := runs(t, "FROM scratch\nRUN apt-get install nano\n")
	require.Len(t, vim, 1)
	require.Len(t, nano, 1)

	// act
	vimKey := plan.RunKey("", vim[0])
	nanoKey := plan.RunKey("", nano[0])

	// assert
	assert.NotEqual(t, vimKey, nanoKey)
}

func TestTheSameRunAfterADifferentStepGetsADifferentKey(t *testing.T) {
	// arrange
	updated := runs(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install vim\n")
	upgraded := runs(t, "FROM scratch\nRUN apt-get upgrade\nRUN apt-get install vim\n")
	require.Len(t, updated, 2)
	require.Len(t, upgraded, 2)

	// act
	updatedKey := plan.RunKey(plan.RunKey("", updated[0]), updated[1])
	upgradedKey := plan.RunKey(plan.RunKey("", upgraded[0]), upgraded[1])

	// assert
	assert.NotEqual(t, updatedKey, upgradedKey)
}
