package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestCheckKeepsItsCommandVerbatim(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCHECK test -f /etc/os-release  &&  echo 'a  b'\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Check{Line: 2, Command: "test -f /etc/os-release  &&  echo 'a  b'"},
	}, stages[0].Instructions)
}
