//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestAFilesTimeIsTheSameWhenEverItWasPlaced(t *testing.T) {
	// arrange
	root := t.TempDir()

	// act
	err := place.Open(root).File("/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(filepath.Join(root, "motd"))
	require.NoError(t, err)
	assert.Equal(t, time.Unix(0, 0).UTC(), info.ModTime().UTC())
}

func TestALinksTimeIsTheSameWhenEverItWasPlaced(t *testing.T) {
	// arrange
	root := t.TempDir()

	// act
	err := place.Open(root).Link("/issue", "motd")

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(filepath.Join(root, "issue"))
	require.NoError(t, err)
	assert.Equal(t, time.Unix(0, 0).UTC(), info.ModTime().UTC())
}
