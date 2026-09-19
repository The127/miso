package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestRunKeepsItsCommandVerbatim(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN echo  'a  b' > /etc/motd\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Line: 2, Command: "echo  'a  b' > /etc/motd"},
	}, stages[0].Instructions)
}

func TestRunWithoutNetworkIsOffline(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN --network=none make test\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Line: 2, Offline: true, Command: "make test"},
	}, stages[0].Instructions)
}

func TestRunWithoutACommandIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.ErrorIs(t, err, imagefile.ErrArguments)
	assert.ErrorContains(t, err, "needs a command")
}
