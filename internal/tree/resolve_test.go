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

func TestASourcePathThroughALinkIsResolvedInsideTheStage(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	// a directory of the VM, which the link names absolutely, and its twin
	// in the stage
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "f"), []byte("host\n"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(stage, outside), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, outside, "f"), []byte("stage\n"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(stage, "x")))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/x/f"}, onto("/f"))

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(image, "f"))
	require.NoError(t, err)
	assert.Equal(t, "stage\n", string(got))
}
