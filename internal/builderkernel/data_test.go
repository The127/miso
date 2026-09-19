package builderkernel_test

import (
	"archive/tar"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulikunitz/xz"

	"github.com/The127/miso/internal/builderkernel"
)

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

func TestTheKernelOfAPackageIsItsBootVmlinuz(t *testing.T) {
	// arrange
	deb := packaged(t, map[string]string{
		"./boot/vmlinuz-6.12.107+deb13-cloud-amd64": "the kernel",
		"./boot/config-6.12.107+deb13-cloud-amd64":  "the config",
	})

	// act
	release, image, err := builderkernel.Kernel(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "6.12.107+deb13-cloud-amd64", release)
	assert.Equal(t, "the kernel", string(image))
}

func TestAPackageWithTwoKernelsIsRefused(t *testing.T) {
	// arrange
	deb := packaged(t, map[string]string{
		"./boot/vmlinuz-6.12.107+deb13-cloud-amd64": "the kernel",
		"./boot/vmlinuz-6.12.108+deb13-cloud-amd64": "another kernel",
	})

	// act
	_, _, err := builderkernel.Kernel(bytes.NewReader(deb))

	// assert
	assert.ErrorContains(t, err, "6.12.107+deb13-cloud-amd64")
	assert.ErrorContains(t, err, "6.12.108+deb13-cloud-amd64")
}
