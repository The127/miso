package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestACheckAndARunOfTheSameCommandGetDifferentKeys(t *testing.T) {
	// arrange
	run := parse(t, "FROM scratch\nOUTPUT disk os.img\nRUN true\n")
	check := parse(t, "FROM scratch\nOUTPUT disk os.img\nCHECK true\n")

	// act
	runKeys := keys(t, run, anyAgent, noFiles, noImages)
	checkKeys := keys(t, check, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, runKeys), lastKey(t, checkKeys))
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
	oneKeys := keys(t, one, anyAgent, files{"a b": "one", "a": "one", "b": "one"}, noImages)
	twoKeys := keys(t, two, anyAgent, files{"a b": "one", "a": "one", "b": "one"}, noImages)

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

func TestAChangedCopySourceChangesItsKey(t *testing.T) {
	// arrange
	app := parse(t, "FROM scratch AS build\nRUN make\nFROM scratch\nCOPY --from=build /out/app /usr/bin/\n")
	tool := parse(t, "FROM scratch AS build\nRUN make\nFROM scratch\nCOPY --from=build /out/tool /usr/bin/\n")

	// act
	appKeys := keys(t, app, anyAgent, noFiles, noImages)
	toolKeys := keys(t, tool, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, appKeys), lastKey(t, toolKeys))
}

func TestAChangedEnvNameChangesTheKeysAfterIt(t *testing.T) {
	// arrange
	editor := parse(t, "FROM scratch\nENV EDITOR=vim\nRUN true\n")
	visual := parse(t, "FROM scratch\nENV VISUAL=vim\nRUN true\n")

	// act
	editorKeys := keys(t, editor, anyAgent, noFiles, noImages)
	visualKeys := keys(t, visual, anyAgent, noFiles, noImages)

	// assert
	assert.NotEqual(t, lastKey(t, editorKeys), lastKey(t, visualKeys))
}
