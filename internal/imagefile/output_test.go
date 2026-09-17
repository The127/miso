package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestOutputNamesItsKindAndItsFile(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nOUTPUT image os.raw\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Output{Line: 2, Kind: "image", Name: "os.raw"},
	}, stages[0].Instructions)
}

func TestOutputWithoutAFileNameIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nOUTPUT image\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "kind and a file name")
}

func TestOutputKeepsItsOptions(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nOUTPUT image os.raw --layout=layouts/os --split\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Output{
			Line:    2,
			Kind:    "image",
			Name:    "os.raw",
			Options: map[string]string{"layout": "layouts/os", "split": ""},
		},
	}, stages[0].Instructions)
}

func TestOutputWithTheSameOptionTwiceIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nOUTPUT image os.raw --size=8G --size=16G\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "--size twice")
}

func TestOutputWithStrayWordsIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nOUTPUT image os.raw extra\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "takes a kind")
}

func TestOutputWithANamelessOptionIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nOUTPUT image os.raw --=16G\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "without a name")
}
