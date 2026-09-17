package buildcontext_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/buildcontext"
)

func TestAMissingPathDoesNotExist(t *testing.T) {
	// arrange
	context := buildcontext.Open(t.TempDir())

	// act
	_, err := context.Digest("nope")

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
}
