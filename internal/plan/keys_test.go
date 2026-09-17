package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

const anyAgent = "agent"

func keys(t *testing.T, stages []imagefile.Stage, agent string, context plan.Context, bases plan.Bases) [][]string {
	t.Helper()

	found, err := plan.Keys(stages, agent, context, bases)
	require.NoError(t, err)

	return found
}

func lastKey(t *testing.T, keys [][]string) string {
	t.Helper()

	require.NotEmpty(t, keys)
	stage := keys[len(keys)-1]
	require.NotEmpty(t, stage)

	return stage[len(stage)-1]
}

func TestAChangedRunCommandChangesItsKey(t *testing.T) {
	// arrange
	vim := parse(t, "FROM scratch\nRUN apt-get install vim\n")
	nano := parse(t, "FROM scratch\nRUN apt-get install nano\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, noImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestTheSameRunAfterADifferentStepGetsADifferentKey(t *testing.T) {
	// arrange
	updated := parse(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install vim\n")
	upgraded := parse(t, "FROM scratch\nRUN apt-get upgrade\nRUN apt-get install vim\n")

	// act
	updatedKeys := keys(t, updated, anyAgent, noFiles, noImages)
	upgradedKeys := keys(t, upgraded, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, updatedKeys), lastKey(t, upgradedKeys))
}

func TestACheckAndARunOfTheSameCommandGetDifferentKeys(t *testing.T) {
	// arrange
	run := parse(t, "FROM scratch\nRUN true\n")
	check := parse(t, "FROM scratch\nCHECK true\n")

	// act
	runKeys := keys(t, run, anyAgent, noFiles, noImages)
	checkKeys := keys(t, check, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, runKeys), lastKey(t, checkKeys))
}

func TestADifferentBaseChangesTheKeys(t *testing.T) {
	// arrange
	sid := parse(t, "FROM debian:sid\nRUN true\n")
	trixie := parse(t, "FROM debian:trixie\nRUN true\n")

	// act
	sidKeys := keys(t, sid, anyAgent, noFiles, noImages)
	trixieKeys := keys(t, trixie, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, sidKeys), lastKey(t, trixieKeys))
}

func TestAChangedEnvChangesTheKeysAfterIt(t *testing.T) {
	// arrange
	vim := parse(t, "FROM scratch\nENV EDITOR=vim\nRUN true\n")
	nano := parse(t, "FROM scratch\nENV EDITOR=nano\nRUN true\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, noImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestAChangedCopyDestinationChangesItsKey(t *testing.T) {
	// arrange
	etc := parse(t, "FROM scratch\nCOPY motd /etc/\n")
	srv := parse(t, "FROM scratch\nCOPY motd /srv/\n")

	// act
	etcKeys := keys(t, etc, anyAgent, files{"motd": "hello"}, noImages)
	srvKeys := keys(t, srv, anyAgent, files{"motd": "hello"}, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, etcKeys), lastKey(t, srvKeys))
}

func TestOneCopySourceWithASpaceIsNotTwoSources(t *testing.T) {
	// arrange
	one := parse(t, "FROM scratch\nCOPY \"a b\" /c\n")
	two := parse(t, "FROM scratch\nCOPY a b /c\n")

	// act
	oneKeys := keys(t, one, anyAgent, files{"a b": "", "a": "", "b": ""}, noImages)
	twoKeys := keys(t, two, anyAgent, files{"a b": "", "a": "", "b": ""}, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, oneKeys), lastKey(t, twoKeys))
}

func TestAChangedOutputKindChangesItsKey(t *testing.T) {
	// arrange
	disk := parse(t, "FROM scratch\nOUTPUT disk os.img\n")
	iso := parse(t, "FROM scratch\nOUTPUT iso os.img\n")

	// act
	diskKeys := keys(t, disk, anyAgent, noFiles, noImages)
	isoKeys := keys(t, iso, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, diskKeys), lastKey(t, isoKeys))
}

func TestAChangedOutputOptionChangesItsKey(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G\n")
	large := parse(t, "FROM scratch\nOUTPUT disk os.img --size=8G\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestOutputOptionsInADifferentOrderKeepTheKey(t *testing.T) {
	// arrange
	sizeFirst := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G --verity --label=root\n")
	sizeLast := parse(t, "FROM scratch\nOUTPUT disk os.img --label=root --verity --size=4G\n")

	// act
	sizeFirstKeys := keys(t, sizeFirst, anyAgent, noFiles, noImages)
	sizeLastKeys := keys(t, sizeLast, anyAgent, noFiles, noImages)

	// assert
	assert.Equal(t, lastKey(t, sizeFirstKeys), lastKey(t, sizeLastKeys))
}

func TestACopyFromAChangedStageGetsADifferentKey(t *testing.T) {
	// arrange
	vim := parse(t, "FROM debian:sid AS build\nRUN make vim\nFROM scratch\nCOPY --from=build /out /\n")
	nano := parse(t, "FROM debian:sid AS build\nRUN make nano\nFROM scratch\nCOPY --from=build /out /\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, noImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestACopyFromTheContextIgnoresTheStagesBeforeIt(t *testing.T) {
	// arrange
	vim := parse(t, "FROM debian:sid\nRUN make vim\nFROM scratch\nCOPY motd /etc/\n")
	nano := parse(t, "FROM debian:sid\nRUN make nano\nFROM scratch\nCOPY motd /etc/\n")

	// act
	vimKeys := keys(t, vim, anyAgent, files{"motd": "hello"}, noImages)
	nanoKeys := keys(t, nano, anyAgent, files{"motd": "hello"}, noImages)

	// assert
	assert.Equal(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
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
