//go:build vmtest

package tree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/tree"
)

// onto lands everything a copy takes at one path of the image.
func onto(path string) tree.Land {
	return func(string, string, bool) (string, error) { return path, nil }
}

func TestAFileOfAStageLandsAtTheDestination(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "etc"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "motd"), []byte("hello\n"), 0o600))
	land := onto("/motd")

	// act
	err := tree.Into(stage, place.Open(image), []string{"/etc/motd"}, land)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(image, "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(got))
}
