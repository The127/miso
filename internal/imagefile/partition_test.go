package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestAPartitionNamesItselfAndKeepsItsSettingsInOrder(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nPARTITION root Format=ext4 SizeMinBytes=3G\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Partition{Line: 2, Name: "root", Settings: []imagefile.Setting{
			{Key: "Format", Value: "ext4"},
			{Key: "SizeMinBytes", Value: "3G"},
		}},
	}, stages[0].Instructions)
}

func TestAPartitionWithoutANameIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nPARTITION\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "needs a name")
}

func TestAPartitionSettingWithoutAValueIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nPARTITION root Format\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "Format")
}
