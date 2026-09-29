package builder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/imagefile"
)

// names are the names of the files in a directory.
func names(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var found []string
	for _, entry := range entries {
		found = append(found, entry.Name())
	}

	return found
}

func TestADiskWithChecksGetsItsNameOnlyAfterTheyPassed(t *testing.T) {
	// arrange
	dir := t.TempDir()
	checks := []imagefile.Check{{Line: 4, Command: "command -v htop"}}
	var booted string
	var there []string
	var ran []imagefile.Check
	check := func(disk string, request build.Request) error {
		content, err := os.ReadFile(disk)
		booted = string(content)
		there = names(t, dir)
		ran = request.Checks

		return err
	}

	// act
	output, err := builder.OutputsIn(dir, check)(build.Request{Output: "os.raw", Checks: checks})
	require.NoError(t, err)
	_, err = output.WriteAt([]byte("disk\n"), 0)
	require.NoError(t, err)
	err = output.Close()

	// assert
	require.NoError(t, err)
	assert.Equal(t, "disk\n", booted)
	assert.NotContains(t, there, "os.raw")
	assert.Equal(t, checks, ran)
	assert.Equal(t, []string{"os.raw"}, names(t, dir))
	written, err := os.ReadFile(filepath.Join(dir, "os.raw"))
	require.NoError(t, err)
	assert.Equal(t, "disk\n", string(written))
}
