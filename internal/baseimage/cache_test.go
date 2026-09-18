package baseimage_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/baseimage"
)

func TestANameNobodyKnowsCannotBeFetched(t *testing.T) {
	// arrange
	cache := baseimage.Open(t.TempDir(), http.DefaultClient, map[string]string{})

	// act
	_, err := cache.Digest("nope")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrUnknownBase)
	assert.ErrorContains(t, err, "nope")
}
