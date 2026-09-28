//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestANewDirectoryHasTheModeOfItsEntry(t *testing.T) {
	// arrange
	root := t.TempDir()
	old := syscall.Umask(0o077)
	t.Cleanup(func() { syscall.Umask(old) })

	// act
	err := place.Open(root).Directory("/srv", 0o1775)

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(filepath.Join(root, "srv"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeDir|os.ModeSticky|0o775, info.Mode())
}
