package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/plan"
)

func TestTwoOutputsWithTheSameFileNameAreRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS one\nOUTPUT image os.raw\nFROM debian:sid AS two\nOUTPUT image os.raw\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrDuplicateOutput)
	assert.ErrorContains(t, err, "os.raw")
}

func TestAnOutputNameThatLeavesTheOutputDirectoryIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT disk ../os.raw\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotAFileName)
	assert.ErrorContains(t, err, "line 2")
}
