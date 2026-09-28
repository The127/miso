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
	found, err := kernel.Find(image)

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
	found, err := kernel.Find(image)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "boot/initrd.img-7.2.8+deb14-amd64", found.Initrd)
}

func TestAKernelWithoutAnInitrdFailsNamingWhereItLooked(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"usr/lib/modules/7.2.8+deb14-amd64/vmlinuz": {},
	}

	// act
	_, err := kernel.Find(image)

	// assert
	require.ErrorIs(t, err, kernel.ErrNoInitrd)
	assert.ErrorContains(t, err, "/usr/lib/modules/7.2.8+deb14-amd64/initrd")
	assert.ErrorContains(t, err, "/boot/initrd.img-7.2.8+deb14-amd64")
}
