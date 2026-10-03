package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

func TestAKernelOutputIsTheKernelOfTheLayersOfItsStageWithoutTools(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nRUN apt-get install htop\nOUTPUT kernel vmlinuz\n")
	require.Len(t, source.Stages, 1)
	image := source.Stages[0]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 4)
	assert.Equal(t, protocol.BootPart{
		Key:    image.Steps[1].Key,
		Layers: []string{image.BaseKey, image.Steps[0].Key},
		Part:   "kernel",
	}, requests[len(requests)-2].Message)
}

func TestAnInitrdOutputIsTheInitrdOfTheLayersOfItsStage(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT initrd initrd.img\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	part, isPart := requests[len(requests)-2].Message.(protocol.BootPart)
	require.True(t, isPart)
	assert.Equal(t, "initrd", part.Part)
}

func TestAKernelOutputWithElfIsUnpackedWithTheLayersOfTheToolsStageMisoAdds(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT kernel --elf vmlinux\n")
	require.Len(t, source.Stages, 2)
	tools, image := source.Stages[0], source.Stages[1]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	assert.Equal(t, protocol.BootPart{
		Key:    image.Steps[0].Key,
		Layers: []string{image.BaseKey},
		Tools:  []string{tools.BaseKey, tools.Steps[0].Key},
		Part:   "kernel",
		ELF:    true,
	}, requests[len(requests)-2].Message)
}

func TestAnElfOptionWithAValueFailsAtItsLineNamingIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT kernel --elf=false vmlinux\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrOptionTakesNoValue)
	assert.ErrorContains(t, err, "line 2")
	assert.ErrorContains(t, err, "--elf")
}

func TestAKernelOutputWithAPathAsksForThatFileOfTheImage(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT kernel --path=/src/linux/arch/x86/boot/bzImage vmlinuz\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	part, isPart := requests[len(requests)-2].Message.(protocol.BootPart)
	require.True(t, isPart)
	assert.Equal(t, "/src/linux/arch/x86/boot/bzImage", part.Path)
}

func TestAnInitrdOutputWithAPathAsksForThatFileOfTheImage(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT initrd --path=/src/initrd.img initrd.img\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	part, isPart := requests[len(requests)-2].Message.(protocol.BootPart)
	require.True(t, isPart)
	assert.Equal(t, "/src/initrd.img", part.Path)
}

func TestAPathThatIsNotAbsoluteFailsAtItsLineNamingIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT kernel --path=arch/x86/boot/bzImage vmlinuz\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrPathNotAbsolute)
	assert.ErrorContains(t, err, "line 2")
	assert.ErrorContains(t, err, "--path")
}
