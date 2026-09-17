package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestFromOpensAStageOnItsBase(t *testing.T) {
	// arrange
	source := "FROM debian:sid\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "debian:sid", stages[0].Base)
}

func TestFromWithAsNamesTheStage(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS build\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "build", stages[0].Name)
}

func TestFromWithoutABaseIsRejected(t *testing.T) {
	// arrange
	source := "FROM\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: FROM needs a base")
	assert.ErrorIs(t, err, imagefile.ErrArguments)
}

func TestFromWithStrayWordsIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid build\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: FROM takes a base and an optional AS name")
}
