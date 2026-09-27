package imagefile_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestAParseErrorKnowsItsLine(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nBOGUS\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	var parseErr *imagefile.Error
	require.ErrorAs(t, err, &parseErr)
	assert.Equal(t, 2, parseErr.Line)
}

func TestAnErrorThatKnowsItsFileReadsFileColonLine(t *testing.T) {
	// arrange
	err := &imagefile.Error{File: "Imagefile", Line: 2, Err: imagefile.ErrNoFrom}

	// act
	said := err.Error()

	// assert
	assert.Equal(t, "Imagefile:2: no FROM instruction", said)
}

func TestAParseErrorInAFileIsNamedAfterIt(t *testing.T) {
	// arrange
	_, parseErr := imagefile.Parse("FROM debian:sid\nBOGUS\n")
	require.Error(t, parseErr)

	// act
	err := imagefile.InFile("Imagefile", parseErr)

	// assert
	assert.EqualError(t, err, "Imagefile:2: unknown instruction BOGUS")
}

func TestAnErrorWithoutALineIsNotNamedAfterAFile(t *testing.T) {
	// arrange
	lineless := errors.New("disk full")

	// act
	err := imagefile.InFile("Imagefile", lineless)

	// assert
	assert.Same(t, lineless, err)
}
