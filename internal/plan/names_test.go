package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/plan"
)

func TestTwoStagesWithTheSameNameAreRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS build\nFROM scratch AS build\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrDuplicateStage)
	assert.ErrorContains(t, err, "build")
}
