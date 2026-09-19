//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestARunCannotLeaveItsRoot(t *testing.T) {
	// arrange
	outside := filepath.Join(t.TempDir(), "outside")
	require.NoError(t, os.WriteFile(outside, []byte("escaped"), 0o600))
	root := onBase(t)
	// the way out of a chroot: chroot deeper, then climb above the old root
	escape := fmt.Sprintf(`perl -e 'mkdir "/away"; chroot "/away" or die $!; chdir ".." for 1..64; chroot "." or die $!; open my $f, "<", "%s" or die $!; print <$f>'`, outside)
	run := protocol.Run{Command: escape}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.NotContains(t, out.String(), "escaped")
}

func TestARunCanChangeItsRootItself(t *testing.T) {
	// arrange
	root := onBase(t)
	// what container tools inside a build do
	pivot := "unshare -m sh -c 'mkdir /new && mount -t tmpfs none /new && mkdir /new/old && pivot_root /new /new/old && echo pivoted'"
	run := protocol.Run{Command: pivot}
	var out bytes.Buffer

	// act
	code, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code, out.String())
	assert.Equal(t, "pivoted\n", out.String())
}
