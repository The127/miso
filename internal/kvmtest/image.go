package kvmtest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/qemu"
)

// Image is a base image miso knows, fetched once into miso's own cache, as
// a disk that forgets what a boot writes.
func Image(t *testing.T, name string) qemu.Disk {
	t.Helper()

	dir := basesDir(t)
	blobs := download.Open(dir, http.DefaultClient)
	bases := baseimage.Open(dir, blobs, baseimage.Known)

	digest, err := bases.Digest(name)
	require.NoError(t, err)
	if digest == "" {
		digest, err = bases.Fetch(t.Context(), name)
		require.NoError(t, err)
	}

	format, err := bases.Format(digest)
	require.NoError(t, err)

	return qemu.Disk{Path: blobs.Path(digest), Format: format, Serial: "image", Access: qemu.Snapshot}
}
