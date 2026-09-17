package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/plan"
)

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
	vimKeys := plan.Keys(vim)
	nanoKeys := plan.Keys(nano)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestTheSameRunAfterADifferentStepGetsADifferentKey(t *testing.T) {
	// arrange
	updated := parse(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install vim\n")
	upgraded := parse(t, "FROM scratch\nRUN apt-get upgrade\nRUN apt-get install vim\n")

	// act
	updatedKeys := plan.Keys(updated)
	upgradedKeys := plan.Keys(upgraded)

	// assert
	assert.NotEqual(t, lastKey(t, updatedKeys), lastKey(t, upgradedKeys))
}

func TestACheckAndARunOfTheSameCommandGetDifferentKeys(t *testing.T) {
	// arrange
	run := parse(t, "FROM scratch\nRUN true\n")
	check := parse(t, "FROM scratch\nCHECK true\n")

	// act
	runKeys := plan.Keys(run)
	checkKeys := plan.Keys(check)

	// assert
	assert.NotEqual(t, lastKey(t, runKeys), lastKey(t, checkKeys))
}

func TestADifferentBaseChangesTheKeys(t *testing.T) {
	// arrange
	sid := parse(t, "FROM debian:sid\nRUN true\n")
	trixie := parse(t, "FROM debian:trixie\nRUN true\n")

	// act
	sidKeys := plan.Keys(sid)
	trixieKeys := plan.Keys(trixie)

	// assert
	assert.NotEqual(t, lastKey(t, sidKeys), lastKey(t, trixieKeys))
}

func TestAChangedEnvChangesTheKeysAfterIt(t *testing.T) {
	// arrange
	vim := parse(t, "FROM scratch\nENV EDITOR=vim\nRUN true\n")
	nano := parse(t, "FROM scratch\nENV EDITOR=nano\nRUN true\n")

	// act
	vimKeys := plan.Keys(vim)
	nanoKeys := plan.Keys(nano)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestAChangedCopyDestinationChangesItsKey(t *testing.T) {
	// arrange
	etc := parse(t, "FROM scratch\nCOPY motd /etc/\n")
	srv := parse(t, "FROM scratch\nCOPY motd /srv/\n")

	// act
	etcKeys := plan.Keys(etc)
	srvKeys := plan.Keys(srv)

	// assert
	assert.NotEqual(t, lastKey(t, etcKeys), lastKey(t, srvKeys))
}

func TestOneCopySourceWithASpaceIsNotTwoSources(t *testing.T) {
	// arrange
	one := parse(t, "FROM scratch\nCOPY \"a b\" /c\n")
	two := parse(t, "FROM scratch\nCOPY a b /c\n")

	// act
	oneKeys := plan.Keys(one)
	twoKeys := plan.Keys(two)

	// assert
	assert.NotEqual(t, lastKey(t, oneKeys), lastKey(t, twoKeys))
}

func TestAChangedOutputKindChangesItsKey(t *testing.T) {
	// arrange
	disk := parse(t, "FROM scratch\nOUTPUT disk os.img\n")
	iso := parse(t, "FROM scratch\nOUTPUT iso os.img\n")

	// act
	diskKeys := plan.Keys(disk)
	isoKeys := plan.Keys(iso)

	// assert
	assert.NotEqual(t, lastKey(t, diskKeys), lastKey(t, isoKeys))
}

func TestAChangedOutputOptionChangesItsKey(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G\n")
	large := parse(t, "FROM scratch\nOUTPUT disk os.img --size=8G\n")

	// act
	smallKeys := plan.Keys(small)
	largeKeys := plan.Keys(large)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestOutputOptionsInADifferentOrderKeepTheKey(t *testing.T) {
	// arrange
	sizeFirst := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G --verity --label=root\n")
	sizeLast := parse(t, "FROM scratch\nOUTPUT disk os.img --label=root --verity --size=4G\n")

	// act
	sizeFirstKeys := plan.Keys(sizeFirst)
	sizeLastKeys := plan.Keys(sizeLast)

	// assert
	assert.Equal(t, lastKey(t, sizeFirstKeys), lastKey(t, sizeLastKeys))
}

func TestACopyFromAChangedStageGetsADifferentKey(t *testing.T) {
	// arrange
	vim := parse(t, "FROM debian:sid AS build\nRUN make vim\nFROM scratch\nCOPY --from=build /out /\n")
	nano := parse(t, "FROM debian:sid AS build\nRUN make nano\nFROM scratch\nCOPY --from=build /out /\n")

	// act
	vimKeys := plan.Keys(vim)
	nanoKeys := plan.Keys(nano)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}
