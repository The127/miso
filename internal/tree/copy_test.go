package tree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/tree"
)

func TestAFileKeepsItsContentAndMode(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "hello"), []byte("hi"), 0o600))
	require.NoError(t, os.Chmod(filepath.Join(source, "hello"), 0o400))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(target, "hello"))
	require.NoError(t, err)
	assert.Equal(t, "hi", string(got))
	info, err := os.Lstat(filepath.Join(target, "hello"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o400), info.Mode())
}
