package builderkernel_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
)

// kept writes bytes into a store by hand and answers their digest, as a
// download of them would have left them there.
func kept(t *testing.T, dir string, content []byte) string {
	t.Helper()

	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sha256"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sha256", digest), content, 0o600))

	return "sha256:" + digest
}

func TestTheBuilderKernelIsReadiedFromThePinnedPackage(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.xz": packedModule(t, "name=btrfs"),
	})
	dir := t.TempDir()
	digest := kept(t, dir, deb)
	store := download.Open(dir, http.DefaultClient)

	// act
	booting, err := builderkernel.Ready(context.Background(), store, "https://snapshot.invalid/linux-image.deb", digest, "btrfs")

	// assert
	require.NoError(t, err)
	assert.Equal(t, testRelease, booting.Release)
	assert.Equal(t, "the kernel", string(booting.Image))
	require.Len(t, booting.Modules, 1)
	assert.Equal(t, "btrfs", booting.Modules[0].Name)
}
