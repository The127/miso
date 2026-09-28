//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestALinkLandsAtItsPathWithItsTargetAsWritten(t *testing.T) {
	// arrange
	root := t.TempDir()

	// act
	err := place.Open(root).Link("/etc/issue", "motd")

	// assert
	require.NoError(t, err)
	target, err := os.Readlink(filepath.Join(root, "etc", "issue"))
	require.NoError(t, err)
	assert.Equal(t, "motd", target)
}

func TestALinkReplacesWhatWasThere(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "motd"), []byte("old\n"), 0o600))

	// act
	err := place.Open(root).Link("/motd", "issue")

	// assert
	require.NoError(t, err)
	target, err := os.Readlink(filepath.Join(root, "motd"))
	require.NoError(t, err)
	assert.Equal(t, "issue", target)
}
