package plan_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func TestADifferentBaseChangesTheKeys(t *testing.T) {
	// arrange
	sid := parse(t, "FROM debian:sid\nRUN true\n")
	trixie := parse(t, "FROM debian:trixie\nRUN true\n")

	// act
	sidKeys := keys(t, sid, anyAgent, noFiles, debianImages)
	trixieKeys := keys(t, trixie, anyAgent, noFiles, debianImages)

	// assert
	assert.NotEqual(t, lastKey(t, sidKeys), lastKey(t, trixieKeys))
}

func TestADifferentAgentChangesTheKeys(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nRUN true\n")

	// act
	oldKeys := keys(t, stages, "agent-a", noFiles, noImages)
	newKeys := keys(t, stages, "agent-b", noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, oldKeys), lastKey(t, newKeys))
}

// images are base images in memory, by what a FROM calls them.
type images map[string]string

var errNoSuchImage = errors.New("no such image")

func (i images) Digest(base string) (string, error) {
	digest, found := i[base]
	if !found {
		return "", errNoSuchImage
	}

	return digest, nil
}

var (
	noImages     = images{}
	debianImages = images{"debian:sid": "sha256:sid", "debian:trixie": "sha256:trixie"}
)

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

func TestAnUnknownBaseIsRejectedAtItsLine(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nFROM nope\n")

	// act
	_, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	var planErr *imagefile.Error
	require.ErrorAs(t, err, &planErr)
	assert.Equal(t, 2, planErr.Line)
	assert.ErrorIs(t, err, errNoSuchImage)
	assert.ErrorContains(t, err, "nope")
}
