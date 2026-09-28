package builderkernel_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/download/downloadtest"
)

func TestTheBuilderKernelIsReadiedFromThePinnedPackage(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.xz": packedModule(t, "name=btrfs"),
	})
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, deb)
	store := download.Open(dir, http.DefaultClient)

	// act
	booting, err := builderkernel.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/linux-image.deb", Digest: digest}, "btrfs")

	// assert
	require.NoError(t, err)
	assert.Equal(t, testRelease, booting.Release)
	assert.Equal(t, "the kernel", string(booting.Image))
	require.Len(t, booting.Modules, 1)
	assert.Equal(t, "btrfs", booting.Modules[0].Name)
}
