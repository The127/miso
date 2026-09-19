package builderkernel_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulikunitz/xz"

	"github.com/The127/miso/internal/builderkernel"
)

// packedModule is a kernel module as a package holds it, xz packed.
func packedModule(t *testing.T, entries ...string) string {
	t.Helper()

	var packed bytes.Buffer

	compressed, err := xz.NewWriter(&packed)
	require.NoError(t, err)

	_, err = compressed.Write(module(t, entries...))
	require.NoError(t, err)
	require.NoError(t, compressed.Close())

	return packed.String()
}

func TestTheModulesOfAPackageAreReadFromWhatItHolds(t *testing.T) {
	// arrange
	deb := packaged(t, map[string]string{
		"./boot/vmlinuz-6.12.107+deb13-cloud-amd64":                            "the kernel",
		"./lib/modules/6.12.107+deb13-cloud-amd64/kernel/fs/btrfs/btrfs.ko.xz": packedModule(t, "depends=libcrc32c", "name=btrfs"),
		"./usr/share/doc/linux-image/changelog":                                "the changelog",
	})

	// act
	found, err := builderkernel.Modules(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "btrfs", found[0].Name)
	assert.Equal(t, []string{"libcrc32c"}, found[0].Depends)
}
