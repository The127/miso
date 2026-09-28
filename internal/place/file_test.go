package place_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestAFileLandsAtItsPathWithItsContent(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "etc"), 0o700))

	// act
	err := place.Open(root).File("/etc/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(root, "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}
