package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func firstRun(t *testing.T, source string) imagefile.Run {
	t.Helper()

	stages := parse(t, source)
	require.Len(t, stages, 1)
	require.NotEmpty(t, stages[0].Instructions)

	run, isRun := stages[0].Instructions[0].(imagefile.Run)
	require.True(t, isRun)

	return run
}

func TestAChangedRunCommandChangesItsKey(t *testing.T) {
	// arrange
	vim := firstRun(t, "FROM scratch\nRUN apt-get install vim\n")
	nano := firstRun(t, "FROM scratch\nRUN apt-get install nano\n")

	// act
	vimKey := plan.RunKey(vim)
	nanoKey := plan.RunKey(nano)

	// assert
	assert.NotEqual(t, vimKey, nanoKey)
}
