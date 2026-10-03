package builder_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/protocol"
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

func TestADiskWhoseChecksFailLeavesNoFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	failing := errors.New("exit code 1")
	check := func(string, build.Request) error { return failing }

	// act
	output, err := builder.OutputsIn(dir, check)(build.Request{Output: "os.raw", Checks: []imagefile.Check{{Line: 4, Command: "false"}}})
	require.NoError(t, err)
	err = output.Close()

	// assert
	assert.ErrorIs(t, err, failing)
	assert.Empty(t, names(t, dir))
}

func TestOnlyCheckedGiveADiskWithoutChecksNoFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	outputs := builder.OnlyChecked(builder.OutputsIn(dir, nil))

	// act
	output, err := outputs(build.Request{Output: "os.raw"})

	// assert
	require.NoError(t, err)
	assert.Nil(t, output)
	assert.Empty(t, names(t, dir))
}

func TestOnlyCheckedGiveADiskWithChecksItsFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	outputs := builder.OnlyChecked(builder.OutputsIn(dir, nil))

	// act
	output, err := outputs(build.Request{Output: "os.raw", Checks: []imagefile.Check{{Line: 4, Command: "true"}}})

	// assert
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, []string{".os.raw.unchecked"}, names(t, dir))
}

func TestOnlyCheckedGiveAFileTheChecksOfALaterOutputNeedItsFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	outputs := builder.OnlyChecked(builder.OutputsIn(dir, nil))

	// act
	output, err := outputs(build.Request{Output: "vmlinuz", Needed: true})

	// assert
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, []string{"vmlinuz"}, names(t, dir))
}

func TestADiskWhoseFetchFailsIsNotBootedAndLeavesNoFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	agent := &sending{disk: "disk\n", err: errors.New("the agent broke off")}
	requests := []build.Request{{Message: protocol.Fetch{Key: "disk"}, Output: "os.raw", Checks: []imagefile.Check{{Line: 4, Command: "true"}}}}
	booted := false
	check := func(string, build.Request) error {
		booted = true

		return nil
	}

	// act
	err := builder.Ask(t.Context(), running{}, dialling(agent), agentName, requests, nil, builder.OutputsIn(dir, check), io.Discard)

	// assert
	require.Error(t, err)
	assert.False(t, booted)
	assert.Empty(t, names(t, dir))
}

func TestADiskWithChecksInADirectoryIsCheckedNextToItsFinalNameAndMovedThere(t *testing.T) {
	// arrange
	dir := t.TempDir()
	check := func(string, build.Request) error { return nil }
	checks := []imagefile.Check{{Line: 4, Command: "true"}}

	// act
	output, err := builder.OutputsIn(dir, check)(build.Request{Output: "updates/os.raw", Checks: checks})
	require.NoError(t, err)
	err = output.Close()

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"updates"}, names(t, dir))
	assert.Equal(t, []string{"os.raw"}, names(t, filepath.Join(dir, "updates")))
}
