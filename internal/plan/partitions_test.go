package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/plan"
)

func TestAPartitionLeavesTheLayersOfTheStepsAfterItAlone(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nPARTITION root Format=ext4\nRUN true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	require.NoError(t, err)
	stage := planned.Stages[0]
	require.Len(t, stage.Steps, 2)
	assert.Equal(t, []string{stage.BaseKey}, stage.Steps[1].BuiltOn)
}

func TestADiskGetsADifferentKeyWhenAPartitionChanges(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nPARTITION root SizeMinBytes=3G\nOUTPUT disk os.img\n")
	large := parse(t, "FROM scratch\nPARTITION root SizeMinBytes=4G\nOUTPUT disk os.img\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, toolsImages)
	largeKeys := keys(t, large, anyAgent, noFiles, toolsImages)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAPartitionNameThatLeavesTheDefinitionsIsRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nPARTITION ../root Format=ext4\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrNotAFileName)
	assert.ErrorContains(t, err, "line 2")
}

func TestTwoPartitionsOfAStageWithTheSameNameAreRejected(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nPARTITION root Format=ext4\nPARTITION root Format=vfat\n")

	// act
	err := plan.Validate(stages)

	// assert
	assert.ErrorIs(t, err, plan.ErrDuplicatePartition)
	assert.ErrorContains(t, err, "line 3")
}
