package builder_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
)

func TestTheBootFilesHoldTheKernelsImage(t *testing.T) {
	// arrange
	kernel := builderkernel.Kernel{Image: []byte("the kernel")}

	// act
	boot, err := builder.WriteBoot(t.TempDir(), kernel, nil)

	// assert
	require.NoError(t, err)
	image, err := os.ReadFile(boot.Kernel)
	require.NoError(t, err)
	assert.Equal(t, "the kernel", string(image))
}
