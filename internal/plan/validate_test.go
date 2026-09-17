package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func parse(t *testing.T, source string) []imagefile.Stage {
	t.Helper()

	stages, err := imagefile.Parse(source)
	require.NoError(t, err)

	return stages
}

func TestAPlanErrorKnowsItsLine(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nRUN true\nCOPY --from=nope a /b\n")

	// act
	err := plan.Validate(stages)

	// assert
	var planErr *imagefile.Error
	require.ErrorAs(t, err, &planErr)
	assert.Equal(t, 3, planErr.Line)
}
