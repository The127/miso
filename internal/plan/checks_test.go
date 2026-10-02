package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/plan"
)

func TestACheckWithoutAnOutputBeforeItIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nCHECK true\nOUTPUT image os.raw\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNothingToCheck)
	assert.ErrorContains(t, err, "line 2")
}

func TestACheckAfterAnOutputIsAccepted(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT image os.raw\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.NoError(t, err)
}

func TestACheckAfterAnOutputThatIsNeverBootedIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT kernel vmlinuz\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotBooted)
	assert.ErrorContains(t, err, "line 3")
	assert.ErrorContains(t, err, "kernel")
}

func TestACheckAfterARootfsIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT rootfs os.ext4\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotBooted)
}
