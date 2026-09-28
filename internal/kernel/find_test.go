package kernel_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/kernel"
)

func TestTheOneKernelOfAnImageIsFoundWithTheInitrdBesideIt(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
		"usr/lib/modules/7.2.8+deb14-amd64/initrd":  {},
	}

	// act
	found, err := kernel.Find(image, "")

	// assert
	require.NoError(t, err)
	assert.Equal(t, kernel.Kernel{
		Version: "7.2.8+deb14-amd64",
		Linux:   "usr/lib/modules/7.2.8+deb14-amd64/vmlinuz",
		Initrd:  "usr/lib/modules/7.2.8+deb14-amd64/initrd",
	}, found)
}

func TestAnInitrdOfDebianIsFoundInBoot(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
		"boot/initrd.img-7.2.8+deb14-amd64":         {},
	}

	// act
	found, err := kernel.Find(image, "")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "boot/initrd.img-7.2.8+deb14-amd64", found.Initrd)
}

func TestModulesLeftOfARemovedKernelAreNoKernel(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.1.0+deb14-amd64/modules.dep": {},
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz":     {},
		"usr/lib/modules/7.2.8+deb14-amd64/initrd":      {},
	}

	// act
	found, err := kernel.Find(image, "")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "7.2.8+deb14-amd64", found.Version)
}

func TestTheNewestOfSeveralKernelsIsFound(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
		"usr/lib/modules/7.2.8+deb14-amd64/initrd":  {},
		"usr/lib/modules/7.3.1+deb14-amd64/vmlinuz": {},
		"usr/lib/modules/7.3.1+deb14-amd64/initrd":  {},
	}

	// act
	found, err := kernel.Find(image, "")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "7.3.1+deb14-amd64", found.Version)
}

func TestAWantedKernelIsFoundOverANewerOne(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
		"usr/lib/modules/7.2.8+deb14-amd64/initrd":  {},
		"usr/lib/modules/7.3.1+deb14-amd64/vmlinuz": {},
		"usr/lib/modules/7.3.1+deb14-amd64/initrd":  {},
	}

	// act
	found, err := kernel.Find(image, "7.2.8+deb14-amd64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "7.2.8+deb14-amd64", found.Version)
}

func TestAWantedKernelThatIsNotInstalledFailsNamingTheInstalledOnes(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
		"usr/lib/modules/7.2.8+deb14-amd64/initrd":  {},
	}

	// act
	_, err := kernel.Find(image, "7.9")

	// assert
	require.ErrorIs(t, err, kernel.ErrNoKernel)
	assert.ErrorContains(t, err, "7.9")
	assert.ErrorContains(t, err, "7.2.8+deb14-amd64")
}

func TestAKernelNameOfTheImageIsQuotedWhenItsInitrdIsMissing(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2\x1b[2J/vmlinuz": {},
	}

	// act
	_, err := kernel.Find(image, "")

	// assert
	require.ErrorIs(t, err, kernel.ErrNoInitrd)
	assert.NotContains(t, err.Error(), "\x1b")
	assert.ErrorContains(t, err, `"7.2\x1b[2J"`)
}

func TestTheKernelNamesOfTheImageAreQuotedWhenTheWantedOneIsMissing(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2\x1b[2J/vmlinuz": {},
		"usr/lib/modules/7.2\x1b[2J/initrd":  {},
	}

	// act
	_, err := kernel.Find(image, "7.9")

	// assert
	require.ErrorIs(t, err, kernel.ErrNoKernel)
	assert.NotContains(t, err.Error(), "\x1b")
	assert.ErrorContains(t, err, `"7.2\x1b[2J"`)
}

func TestAnImageWithoutModulesHasNoKernel(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"etc/os-release": {},
	}

	// act
	_, err := kernel.Find(image, "")

	// assert
	require.ErrorIs(t, err, kernel.ErrNoKernel)
	assert.ErrorContains(t, err, "/usr/lib/modules/*/vmlinuz")
}

func TestAnImageWithOnlyModulesLeftHasNoKernel(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.1.0+deb14-amd64/modules.dep": {},
	}

	// act
	_, err := kernel.Find(image, "")

	// assert
	require.ErrorIs(t, err, kernel.ErrNoKernel)
	assert.ErrorContains(t, err, "/usr/lib/modules/*/vmlinuz")
}

func TestAnInitrdOfFedoraIsFoundInBoot(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/6.17.1-300.fc43.x86_64/vmlinuz": {},
		"boot/initramfs-6.17.1-300.fc43.x86_64.img":      {},
	}

	// act
	found, err := kernel.Find(image, "")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "boot/initramfs-6.17.1-300.fc43.x86_64.img", found.Initrd)
}

func TestAKernelWithoutAnInitrdFailsNamingWhereItLooked(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
	}

	// act
	_, err := kernel.Find(image, "")

	// assert
	require.ErrorIs(t, err, kernel.ErrNoInitrd)
	assert.ErrorContains(t, err, "/usr/lib/modules/7.2.8+deb14-amd64/initrd")
	assert.ErrorContains(t, err, "/boot/initrd.img-7.2.8+deb14-amd64")
}
