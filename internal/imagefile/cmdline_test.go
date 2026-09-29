package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestACmdlineKeepsItsTextAsWritten(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCMDLINE rw console=ttyS0 quiet=\"a  b\"\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Cmdline{Line: 2, Text: "rw console=ttyS0 quiet=\"a  b\""},
	}, stages[0].Instructions)
}

func TestACmdlineWithoutTextIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCMDLINE\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "needs some text")
}
