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
