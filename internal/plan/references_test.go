package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/plan"
)

func TestCopyFromAnUnknownStageIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nCOPY --from=nope a /b\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrUnknownStage)
	assert.ErrorContains(t, err, "nope")
}

func TestCopyFromAnEarlierStageIsAccepted(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS build\nFROM scratch\nCOPY --from=build /out/app /usr/bin/\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.NoError(t, err)
}
