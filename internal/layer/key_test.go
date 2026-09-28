package layer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/layer"
)

func TestAKeyThatIsNoPlainNameIsRefused(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())

	// act
	_, err := store.Has("..")

	// assert
	assert.ErrorIs(t, err, layer.ErrBadKey)
}
