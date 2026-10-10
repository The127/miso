package boot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/boot"
)

func TestAnArm64EmulatorIsRegisteredAsDebianRegistersQEMUForAArch64(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "arm64"), []byte("qemu"), 0o600))
	var written []string

	// act
	err := boot.Register(dir, func(entry string) error {
		written = append(written, entry)

		return nil
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{
		`:arm64:M::\x7f\x45\x4c\x46\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:` + filepath.Join(dir, "arm64-binfmt-P") + `:OPF`,
	}, written)
}

func TestAnAmd64EmulatorIsRegisteredAsDebianRegistersQEMUForX86_64(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "amd64"), []byte("qemu"), 0o600))
	var written []string

	// act
	err := boot.Register(dir, func(entry string) error {
		written = append(written, entry)

		return nil
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{
		`:amd64:M::\x7f\x45\x4c\x46\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x3e\x00:\xff\xff\xff\xff\xff\xfe\xfe\xfc\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:` + filepath.Join(dir, "amd64-binfmt-P") + `:OPF`,
	}, written)
}
