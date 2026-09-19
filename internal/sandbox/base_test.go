//go:build vmtest

package sandbox_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/basemount"
	"github.com/The127/miso/internal/sandbox"
)

// onBase is a root for a run, the VM's base image with what the run writes
// going above it and the floor below it, as the agent stacks its layers.
func onBase(t *testing.T) string {
	t.Helper()

	require.NotEmpty(t, os.Getenv("MISO_VMTEST_BASE_DIGEST"), "MISO_VMTEST_BASE names no base image")

	scratch := t.TempDir()
	base := filepath.Join(scratch, "base")
	upper := filepath.Join(scratch, "upper")
	work := filepath.Join(scratch, "work")
	root := filepath.Join(scratch, "root")

	for _, dir := range []string{base, upper, work, root} {
		require.NoError(t, os.Mkdir(dir, 0o700))
	}

	require.NoError(t, basemount.Mount("miso-test-base", base))
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(base, syscall.MNT_DETACH)) })

	floor, err := sandbox.Floor(scratch)
	require.NoError(t, err)

	options := "lowerdir=" + base + ":" + floor + ",upperdir=" + upper + ",workdir=" + work
	require.NoError(t, syscall.Mount("overlay", root, "overlay", 0, options))
	// before the scratch is removed, which must never reach into the root
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(root, syscall.MNT_DETACH)) })

	return root
}
