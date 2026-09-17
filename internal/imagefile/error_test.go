package imagefile_test

import (
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
