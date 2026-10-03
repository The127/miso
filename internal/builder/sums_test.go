package builder_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
)

// sumOf is the line of SHA256SUMS for content under a name.
func sumOf(name, content string) string {
	return fmt.Sprintf("%x  %s\n", sha256.Sum256([]byte(content)), name)
}

func TestAnOutputThatIsListedGetsItsSumInTheFileOfSumsOfItsDirectory(t *testing.T) {
	// arrange
	dir := t.TempDir()
	output, err := builder.OutputsIn(dir, nil)(build.Request{Output: "updates/root_1.2.raw", Listed: true})
	require.NoError(t, err)
	_, err = output.WriteAt([]byte("root\n"), 0)
	require.NoError(t, err)

	// act
	err = output.Close()

	// assert
	require.NoError(t, err)
	sums, err := os.ReadFile(filepath.Join(dir, "updates", "SHA256SUMS"))
	require.NoError(t, err)
	assert.Equal(t, sumOf("root_1.2.raw", "root\n"), string(sums))
}

func TestTheSumsOfListedOutputsOfADirectoryAreInTheOrderOfTheirNames(t *testing.T) {
	// arrange
	dir := t.TempDir()
	for _, file := range []struct{ name, content string }{{"updates/uki_1.2.efi", "uki\n"}, {"updates/root_1.2.raw", "root\n"}} {
		output, err := builder.OutputsIn(dir, nil)(build.Request{Output: file.name, Listed: true})
		require.NoError(t, err)
		_, err = output.WriteAt([]byte(file.content), 0)
		require.NoError(t, err)
		require.NoError(t, output.Close())
	}

	// act
	sums, err := os.ReadFile(filepath.Join(dir, "updates", "SHA256SUMS"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, sumOf("root_1.2.raw", "root\n")+sumOf("uki_1.2.efi", "uki\n"), string(sums))
}

func TestAnOutputThatIsNotListedGetsNoFileOfSums(t *testing.T) {
	// arrange
	dir := t.TempDir()
	output, err := builder.OutputsIn(dir, nil)(build.Request{Output: "updates/root_1.2.raw"})
	require.NoError(t, err)

	// act
	err = output.Close()

	// assert
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(dir, "updates", "SHA256SUMS"))
}

func TestAListedOutputBuiltAgainReplacesItsOldSum(t *testing.T) {
	// arrange
	dir := t.TempDir()
	for _, content := range []string{"old\n", "new\n"} {
		output, err := builder.OutputsIn(dir, nil)(build.Request{Output: "updates/root_1.2.raw", Listed: true})
		require.NoError(t, err)
		_, err = output.WriteAt([]byte(content), 0)
		require.NoError(t, err)
		require.NoError(t, output.Close())
	}

	// act
	sums, err := os.ReadFile(filepath.Join(dir, "updates", "SHA256SUMS"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, sumOf("root_1.2.raw", "new\n"), string(sums))
}
