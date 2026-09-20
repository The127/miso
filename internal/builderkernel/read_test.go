package builderkernel_test

import (
	"archive/tar"
	"bytes"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulikunitz/xz"

	"github.com/The127/miso/internal/builderkernel"
)

// testRelease is the kernel release the fixtures are built around.
const testRelease = "6.12.107+deb13-cloud-amd64"

// packaged is a Debian package whose data holds the files, as the kernel
// package holds its kernel and modules.
func packaged(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var packed bytes.Buffer

	compressed, err := xz.NewWriter(&packed)
	require.NoError(t, err)

	writer := tar.NewWriter(compressed)
	for name, content := range files {
		require.NoError(t, writer.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}))
		_, err := writer.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())
	require.NoError(t, compressed.Close())

	return archive(t, member{"debian-binary", "2.0\n"}, member{"data.tar.xz", packed.String()})
}

// wholePackage is a package that holds everything miso reads out of one, so
// that a test names only the files it is about.
func wholePackage(t *testing.T, files map[string]string) []byte {
	t.Helper()

	whole := map[string]string{
		"./boot/vmlinuz-" + testRelease:                     "the kernel",
		"./lib/modules/" + testRelease + "/modules.builtin": "kernel/drivers/nothing.ko\n",
	}

	maps.Copy(whole, files)

	return packaged(t, whole)
}

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

func TestTheKernelOfAPackageIsItsBootVmlinuz(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./boot/vmlinuz-" + testRelease: "the kernel",
		"./boot/config-" + testRelease:  "the config",
	})

	// act
	held, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	assert.Equal(t, testRelease, held.Release)
	assert.Equal(t, "the kernel", string(held.Image))
}

func TestAPackageWithTwoKernelsIsRefused(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./boot/vmlinuz-" + testRelease:             "the kernel",
		"./boot/vmlinuz-6.12.108+deb13-cloud-amd64": "another kernel",
	})

	// act
	_, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	assert.ErrorContains(t, err, testRelease)
	assert.ErrorContains(t, err, "6.12.108+deb13-cloud-amd64")
}

func TestAPackageWithNoKernelIsRefused(t *testing.T) {
	// arrange
	deb := packaged(t, map[string]string{"./usr/share/doc/linux-image/changelog": "the changelog"})

	// act
	_, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	assert.ErrorContains(t, err, "no kernel")
}

func TestTheModulesOfAPackageAreReadFromWhatItHolds(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.xz": packedModule(t, "depends=libcrc32c", "name=btrfs"),
		"./usr/share/doc/linux-image/changelog":                         "the changelog",
	})

	// act
	held, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	require.Len(t, held.Modules, 1)
	assert.Equal(t, "btrfs", held.Modules[0].Name)
	assert.Equal(t, []string{"libcrc32c"}, held.Modules[0].Depends)
}

func TestAModulePackedOtherwiseIsRefused(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.zst": "packed some other way",
	})

	// act
	_, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	assert.ErrorContains(t, err, "btrfs.ko.zst")
	assert.ErrorContains(t, err, "btrfs.ko.xz")
}

func TestWhatTheKernelBuildsInIsNamedByItsModulesBuiltin(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/modules.builtin": "kernel/fs/btrfs/btrfs.ko\nkernel/crypto/xor.ko\n",
	})

	// act
	held, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"btrfs", "xor"}, held.Builtin)
}

func TestAPackageWithNoModulesBuiltinIsRefused(t *testing.T) {
	// arrange
	deb := packaged(t, map[string]string{"./boot/vmlinuz-" + testRelease: "the kernel"})

	// act
	_, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	assert.ErrorContains(t, err, "modules.builtin")
}

func TestAPackageCarriesTheBytesOfEachModule(t *testing.T) {
	// arrange
	packed := packedModule(t, "name=btrfs")
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.xz": packed,
	})

	// act
	held, err := builderkernel.Read(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	require.Len(t, held.Modules, 1)
	assert.Equal(t, packed, string(held.Modules[0].Packed))
}
