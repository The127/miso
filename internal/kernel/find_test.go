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
