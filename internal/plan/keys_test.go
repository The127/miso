package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func stage(t *testing.T, source string) imagefile.Stage {
	t.Helper()

	stages := parse(t, source)
	require.Len(t, stages, 1)

	return stages[0]
}

func TestAChangedRunCommandChangesItsKey(t *testing.T) {
	// arrange
	vim := stage(t, "FROM scratch\nRUN apt-get install vim\n")
	nano := stage(t, "FROM scratch\nRUN apt-get install nano\n")

	// act
	vimKeys := plan.Keys(vim)
	nanoKeys := plan.Keys(nano)

	// assert
	require.Len(t, vimKeys, 1)
	require.Len(t, nanoKeys, 1)
	assert.NotEqual(t, vimKeys[0], nanoKeys[0])
}

func TestTheSameRunAfterADifferentStepGetsADifferentKey(t *testing.T) {
	// arrange
	updated := stage(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install vim\n")
	upgraded := stage(t, "FROM scratch\nRUN apt-get upgrade\nRUN apt-get install vim\n")

	// act
	updatedKeys := plan.Keys(updated)
	upgradedKeys := plan.Keys(upgraded)

	// assert
	require.Len(t, updatedKeys, 2)
	require.Len(t, upgradedKeys, 2)
	assert.NotEqual(t, updatedKeys[1], upgradedKeys[1])
}

func TestACheckAndARunOfTheSameCommandGetDifferentKeys(t *testing.T) {
	// arrange
	run := stage(t, "FROM scratch\nRUN true\n")
	check := stage(t, "FROM scratch\nCHECK true\n")

	// act
	runKeys := plan.Keys(run)
	checkKeys := plan.Keys(check)

	// assert
	require.Len(t, runKeys, 1)
	require.Len(t, checkKeys, 1)
	assert.NotEqual(t, runKeys[0], checkKeys[0])
}

func TestADifferentBaseChangesTheKeys(t *testing.T) {
	// arrange
	sid := stage(t, "FROM debian:sid\nRUN true\n")
	trixie := stage(t, "FROM debian:trixie\nRUN true\n")

	// act
	sidKeys := plan.Keys(sid)
	trixieKeys := plan.Keys(trixie)

	// assert
	require.Len(t, sidKeys, 1)
	require.Len(t, trixieKeys, 1)
	assert.NotEqual(t, sidKeys[0], trixieKeys[0])
}

func TestAChangedEnvChangesTheKeysAfterIt(t *testing.T) {
	// arrange
	vim := stage(t, "FROM scratch\nENV EDITOR=vim\nRUN true\n")
	nano := stage(t, "FROM scratch\nENV EDITOR=nano\nRUN true\n")

	// act
	vimKeys := plan.Keys(vim)
	nanoKeys := plan.Keys(nano)

	// assert
	require.Len(t, vimKeys, 2)
	require.Len(t, nanoKeys, 2)
	assert.NotEqual(t, vimKeys[1], nanoKeys[1])
}

func TestAChangedCopyDestinationChangesItsKey(t *testing.T) {
	// arrange
	etc := stage(t, "FROM scratch\nCOPY motd /etc/\n")
	srv := stage(t, "FROM scratch\nCOPY motd /srv/\n")

	// act
	etcKeys := plan.Keys(etc)
	srvKeys := plan.Keys(srv)

	// assert
	require.Len(t, etcKeys, 1)
	require.Len(t, srvKeys, 1)
	assert.NotEqual(t, etcKeys[0], srvKeys[0])
}

func TestOneCopySourceWithASpaceIsNotTwoSources(t *testing.T) {
	// arrange
	one := stage(t, "FROM scratch\nCOPY \"a b\" /c\n")
	two := stage(t, "FROM scratch\nCOPY a b /c\n")

	// act
	oneKeys := plan.Keys(one)
	twoKeys := plan.Keys(two)

	// assert
	require.Len(t, oneKeys, 1)
	require.Len(t, twoKeys, 1)
	assert.NotEqual(t, oneKeys[0], twoKeys[0])
}

func TestAChangedOutputKindChangesItsKey(t *testing.T) {
	// arrange
	disk := stage(t, "FROM scratch\nOUTPUT disk os.img\n")
	iso := stage(t, "FROM scratch\nOUTPUT iso os.img\n")

	// act
	diskKeys := plan.Keys(disk)
	isoKeys := plan.Keys(iso)

	// assert
	require.Len(t, diskKeys, 1)
	require.Len(t, isoKeys, 1)
	assert.NotEqual(t, diskKeys[0], isoKeys[0])
}
