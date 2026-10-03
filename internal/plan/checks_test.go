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

func TestACheckAfterAPortableIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT portable app.raw\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotBooted)
}

func TestACheckAfterASysextIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT sysext app.raw\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotBooted)
}

func TestACheckAfterAConfextIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT confext app.raw\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotBooted)
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

func TestACheckAfterARootfsWithAKernelAndAnInitrdAboveItIsAccepted(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nCMDLINE root=/dev/vda rw\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.NoError(t, err)
}

func TestACheckAfterARootfsWithoutAKernelAboveItIsRejectedNamingTheKernel(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNoBootFiles)
	assert.ErrorContains(t, err, "line 4")
	assert.ErrorContains(t, err, "OUTPUT kernel")
}

func TestACheckAfterARootfsWithoutAnInitrdAboveItIsRejectedNamingTheInitrd(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT kernel vmlinuz\nOUTPUT rootfs os.ext4\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNoBootFiles)
	assert.ErrorContains(t, err, "OUTPUT initrd")
}

func TestAKernelBelowARootfsDoesNotBootIt(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nOUTPUT kernel vmlinuz\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotBooted)
}

func TestACheckAfterARootfsWithoutARootInTheCmdlineIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nCMDLINE rw console=ttyS0\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNoRoot)
	assert.ErrorContains(t, err, "line 6")
	assert.ErrorContains(t, err, "root=")
}

func TestARootfsTypeOnTheCmdlineIsNoRootForTheCheck(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nCMDLINE rootfstype=ext4 rw\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNoRoot)
}

func TestARootOnACmdlineBelowTheRootfsIsNoRootForTheCheck(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nCMDLINE root=/dev/vda\nCHECK true\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNoRoot)
}
