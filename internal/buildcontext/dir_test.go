package buildcontext_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
)

func open(t *testing.T, dir string) *buildcontext.Dir {
	t.Helper()

	context, err := buildcontext.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, context.Close()) })

	return context
}

func TestABuildContextThatIsNotThereCannotBeOpened(t *testing.T) {
	// arrange
	missing := filepath.Join(t.TempDir(), "nope")

	// act
	_, err := buildcontext.Open(missing)

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
}
