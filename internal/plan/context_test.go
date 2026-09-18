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

func TestACopyOfAMissingFileIsRejectedAtItsLine(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCOPY nope /etc/\n")

	// act
	_, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	var planErr *imagefile.Error
	require.ErrorAs(t, err, &planErr)
	assert.Equal(t, 2, planErr.Line)
	assert.ErrorIs(t, err, fs.ErrNotExist)
	assert.ErrorContains(t, err, "nope")
}

func TestACopyOnAnUnfetchedBaseStillListsItsFiles(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nCOPY motd /etc/\n")

	// act
	planned, err := plan.New(stages, anyAgent, files{"motd": "hello"}, images{"debian:sid": ""})

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	require.Len(t, planned.Stages[0].Steps, 1)
	assert.Equal(t, []plan.File{{Path: "motd", Digest: "hello"}}, planned.Stages[0].Steps[0].Files)
	assert.Empty(t, planned.Stages[0].Steps[0].Key)
}
