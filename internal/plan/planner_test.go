package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/plan"
)

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

func TestACopyFromAChangedStageGetsADifferentKey(t *testing.T) {
	// arrange
	vim := parse(t, "FROM debian:sid AS build\nRUN make vim\nFROM scratch\nCOPY --from=build /out /\n")
	nano := parse(t, "FROM debian:sid AS build\nRUN make nano\nFROM scratch\nCOPY --from=build /out /\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, debianImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, debianImages)

	// assert
	assert.NotEqual(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestACopyFromTheContextIgnoresTheStagesBeforeIt(t *testing.T) {
	// arrange
	vim := parse(t, "FROM debian:sid\nRUN make vim\nFROM scratch\nCOPY motd /etc/\n")
	nano := parse(t, "FROM debian:sid\nRUN make nano\nFROM scratch\nCOPY motd /etc/\n")

	// act
	vimKeys := keys(t, vim, anyAgent, files{"motd": "hello"}, debianImages)
	nanoKeys := keys(t, nano, anyAgent, files{"motd": "hello"}, debianImages)

	// assert
	assert.Equal(t, lastKey(t, vimKeys), lastKey(t, nanoKeys))
}

func TestAStepKeepsItsKeyWhenALaterStepChanges(t *testing.T) {
	// arrange
	vim := parse(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install vim\n")
	nano := parse(t, "FROM scratch\nRUN apt-get update\nRUN apt-get install nano\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, noImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, noImages)

	// assert
	require.Len(t, vimKeys, 1)
	require.Len(t, nanoKeys, 1)
	require.Len(t, vimKeys[0], 2)
	require.Len(t, nanoKeys[0], 2)
	assert.Equal(t, vimKeys[0][0], nanoKeys[0][0])
	assert.NotEqual(t, vimKeys[0][1], nanoKeys[0][1])
}

func TestAStageKeepsItsKeysWhenALaterStageChanges(t *testing.T) {
	// arrange
	vim := parse(t, "FROM scratch AS build\nRUN make\nFROM scratch\nRUN apt-get install vim\n")
	nano := parse(t, "FROM scratch AS build\nRUN make\nFROM scratch\nRUN apt-get install nano\n")

	// act
	vimKeys := keys(t, vim, anyAgent, noFiles, noImages)
	nanoKeys := keys(t, nano, anyAgent, noFiles, noImages)

	// assert
	require.Len(t, vimKeys, 2)
	require.Len(t, nanoKeys, 2)
	assert.Equal(t, vimKeys[0], nanoKeys[0])
	assert.NotEqual(t, vimKeys[1], nanoKeys[1])
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

func TestAStageBasedOnAnEmptyStageStillStartsOnItsBase(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS base\nFROM base\nRUN true\n")

	// act
	oldKeys := keys(t, stages, anyAgent, noFiles, images{"debian:sid": "sha256:old"})
	newKeys := keys(t, stages, anyAgent, noFiles, images{"debian:sid": "sha256:new"})

	// assert
	assert.NotEqual(t, lastKey(t, oldKeys), lastKey(t, newKeys))
}

func TestACopyFromAStageDoesNotLookInTheContext(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS build\nRUN make\nFROM scratch\nCOPY --from=build /out /\n")

	// act
	withoutKeys := keys(t, stages, anyAgent, noFiles, debianImages)
	withKeys := keys(t, stages, anyAgent, files{"/out": "unrelated"}, debianImages)

	// assert
	assert.Equal(t, lastKey(t, withoutKeys), lastKey(t, withKeys))
}

func TestAChangedOutputChangesTheKeyOfTheCheckAfterIt(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G\nCHECK true\n")
	large := parse(t, "FROM scratch\nOUTPUT disk os.img --size=8G\nCHECK true\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAChangedEarlierOutputChangesTheKeyOfTheCheck(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G\nOUTPUT portable app.raw\nCHECK true\n")
	large := parse(t, "FROM scratch\nOUTPUT disk os.img --size=8G\nOUTPUT portable app.raw\nCHECK true\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAChangedCheckChangesTheKeyOfTheCheckAfterIt(t *testing.T) {
	// arrange
	running := parse(t, "FROM scratch\nOUTPUT disk os.img\nCHECK systemctl is-system-running\nCHECK test -e /etc/motd\n")
	degraded := parse(t, "FROM scratch\nOUTPUT disk os.img\nCHECK systemctl --failed\nCHECK test -e /etc/motd\n")

	// act
	runningKeys := keys(t, running, anyAgent, noFiles, noImages)
	degradedKeys := keys(t, degraded, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, runningKeys), lastKey(t, degradedKeys))
}

func TestACopyOfAChangedOutputGetsADifferentKey(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch AS vmhost\nRUN make\nOUTPUT portable vmhost.raw --size=1G\nFROM scratch\nCOPY --from=vmhost vmhost.raw /var/components/\n")
	large := parse(t, "FROM scratch AS vmhost\nRUN make\nOUTPUT portable vmhost.raw --size=2G\nFROM scratch\nCOPY --from=vmhost vmhost.raw /var/components/\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAChangedOutputKeepsTheKeyOfTheRunAfterIt(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G\nRUN apt-get install -y vim\n")
	large := parse(t, "FROM scratch\nOUTPUT disk os.img --size=8G\nRUN apt-get install -y vim\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.Equal(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestACopyOfAChangedEarlierOutputGetsADifferentKey(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch AS vmhost\nOUTPUT portable vmhost.raw --size=1G\nOUTPUT portable tools.raw\nFROM scratch\nCOPY --from=vmhost vmhost.raw /var/components/\n")
	large := parse(t, "FROM scratch AS vmhost\nOUTPUT portable vmhost.raw --size=2G\nOUTPUT portable tools.raw\nFROM scratch\nCOPY --from=vmhost vmhost.raw /var/components/\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAChangedOutputKeepsTheKeyOfTheOutputAfterIt(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch\nOUTPUT disk os.img --size=4G\nOUTPUT portable app.raw\n")
	large := parse(t, "FROM scratch\nOUTPUT disk os.img --size=8G\nOUTPUT portable app.raw\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.Equal(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAChangedCheckKeepsTheKeyOfTheRunAfterIt(t *testing.T) {
	// arrange
	running := parse(t, "FROM scratch\nOUTPUT disk os.img\nCHECK systemctl is-system-running\nRUN apt-get install -y vim\n")
	degraded := parse(t, "FROM scratch\nOUTPUT disk os.img\nCHECK systemctl --failed\nRUN apt-get install -y vim\n")

	// act
	runningKeys := keys(t, running, anyAgent, noFiles, noImages)
	degradedKeys := keys(t, degraded, anyAgent, noFiles, noImages)

	// assert
	assert.Equal(t, lastKey(t, runningKeys), lastKey(t, degradedKeys))
}

func TestACopyOfAnOutputKeepsItsKeyWhenAnotherOutputChanges(t *testing.T) {
	// arrange
	small := parse(t, "FROM scratch AS vmhost\nOUTPUT portable vmhost.raw\nOUTPUT portable tools.raw --size=1G\nFROM scratch\nCOPY --from=vmhost vmhost.raw /var/components/\n")
	large := parse(t, "FROM scratch AS vmhost\nOUTPUT portable vmhost.raw\nOUTPUT portable tools.raw --size=2G\nFROM scratch\nCOPY --from=vmhost vmhost.raw /var/components/\n")

	// act
	smallKeys := keys(t, small, anyAgent, noFiles, noImages)
	largeKeys := keys(t, large, anyAgent, noFiles, noImages)

	// assert
	assert.Equal(t, lastKey(t, smallKeys), lastKey(t, largeKeys))
}

func TestAChangedCheckKeepsTheKeysOfAStageBasedOnItsStage(t *testing.T) {
	// arrange
	running := parse(t, "FROM scratch AS os\nOUTPUT disk os.img\nCHECK systemctl is-system-running\nFROM os\nRUN apt-get install -y vim\n")
	degraded := parse(t, "FROM scratch AS os\nOUTPUT disk os.img\nCHECK systemctl --failed\nFROM os\nRUN apt-get install -y vim\n")

	// act
	runningKeys := keys(t, running, anyAgent, noFiles, noImages)
	degradedKeys := keys(t, degraded, anyAgent, noFiles, noImages)

	// assert
	assert.Equal(t, lastKey(t, runningKeys), lastKey(t, degradedKeys))
}

func TestAStageOnAnUnfetchedStageGetsNoKeys(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS bootstrap\nFROM bootstrap\nRUN true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, images{"debian:sid": ""})

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 2)
	require.Len(t, planned.Stages[1].Steps, 1)
	assert.Empty(t, planned.Stages[1].Steps[0].Key)
}

func TestACopyFromAnUnfetchedStageGetsNoKeyWhileTheStepsBeforeItKeepTheirs(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS bootstrap\nFROM scratch\nRUN true\nCOPY --from=bootstrap /rootfs/ /\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, images{"debian:sid": ""})

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 2)
	require.Len(t, planned.Stages[1].Steps, 2)
	assert.NotEmpty(t, planned.Stages[1].Steps[0].Key)
	assert.Empty(t, planned.Stages[1].Steps[1].Key)
}
