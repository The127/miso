package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// images are base images in memory, by what a FROM calls them.
type images map[string]string

func (i images) Digest(base string) string {
	return i[base]
}

var noImages = images{}

func TestARefreshedBaseImageChangesTheKeys(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nRUN true\n")

	// act
	oldKeys := keys(t, stages, anyAgent, noFiles, images{"debian:sid": "sha256:old"})
	newKeys := keys(t, stages, anyAgent, noFiles, images{"debian:sid": "sha256:new"})

	// assert
	assert.NotEqual(t, lastKey(t, oldKeys), lastKey(t, newKeys))
}

func TestScratchIsNotLookedUp(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nRUN true\n")

	// act
	withoutKeys := keys(t, stages, anyAgent, noFiles, noImages)
	withKeys := keys(t, stages, anyAgent, noFiles, images{"scratch": "sha256:unrelated"})

	// assert
	assert.Equal(t, lastKey(t, withoutKeys), lastKey(t, withKeys))
}

func TestAStageBasedOnAChangedStageGetsDifferentKeys(t *testing.T) {
	// arrange
	vim := parse(t, "FROM scratch AS build\nRUN make vim\nFROM build\nRUN make install\n")
	nano := parse(t, "FROM scratch AS build\nRUN make nano\nFROM build\nRUN make install\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, noImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}
