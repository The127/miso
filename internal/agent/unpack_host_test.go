package agent_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/agent"
)

func TestAnELFOfAKernelThatIsNoBzImageIsRefusedNamingTheArm64Image(t *testing.T) {
	// arrange
	image := fstest.MapFS{"boot/vmlinuz": {Data: []byte("MZ an arm64 Image, ARMd")}}

	// act
	_, _, err := agent.KeepStream(image, "boot/vmlinuz", t.TempDir())

	// assert
	assert.ErrorContains(t, err, "--elf")
	assert.ErrorContains(t, err, "arm64")
}
