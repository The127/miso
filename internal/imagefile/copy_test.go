package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestCopyNamesItsSourcesAndItsDestination(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY go.mod go.sum /src/\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Copy{Sources: []string{"go.mod", "go.sum"}, Destination: "/src/"},
	}, stages[0].Instructions)
}

func TestCopyWithoutADestinationIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY onlyone\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: COPY needs a source and a destination")
}
