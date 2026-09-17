package plan_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"

	"github.com/The127/miso/internal/plan"
)

// files is a context in memory, a file's content stands in for its digest.
type files map[string]string

func (f files) Digest(path string) (string, error) {
	content, found := f[path]
	if !found {
		return "", fs.ErrNotExist
	}

	return content, nil
}

var noFiles = files{}

func TestAChangedContextFileChangesTheKeyOfItsCopy(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCOPY motd /etc/\n")

	// act
	helloKeys := keys(t, stages, anyAgent, files{"motd": "hello"}, noImages)
	goodbyeKeys := keys(t, stages, anyAgent, files{"motd": "goodbye"}, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, helloKeys), lastKey(t, goodbyeKeys))
}

func TestACopyFromAStageDoesNotLookInTheContext(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS build\nRUN make\nFROM scratch\nCOPY --from=build /out /\n")

	// act
	withoutKeys := keys(t, stages, anyAgent, noFiles, noImages)
	withKeys := keys(t, stages, anyAgent, files{"/out": "unrelated"}, noImages)

	// assert
	assert.Equal(t, lastKey(t, withoutKeys), lastKey(t, withKeys))
}

func TestACopyOfAMissingFileIsRejectedAtItsLine(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCOPY nope /etc/\n")

	// act
	_, err := plan.Keys(stages, anyAgent, noFiles, noImages)

	// assert
	var planErr *imagefile.Error
	require.ErrorAs(t, err, &planErr)
	assert.Equal(t, 2, planErr.Line)
	assert.ErrorIs(t, err, fs.ErrNotExist)
	assert.ErrorContains(t, err, "nope")
}
