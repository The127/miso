//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestARunHasANullDevice(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "test -c /dev/null"}

	// act
	code, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestARunHasTheUsualDevicesForEveryone(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "stat -c '%n %F %a %t:%T' /dev/null /dev/zero /dev/full /dev/random /dev/urandom /dev/tty"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "/dev/null character special file 666 1:3\n"+
		"/dev/zero character special file 666 1:5\n"+
		"/dev/full character special file 666 1:7\n"+
		"/dev/random character special file 666 1:8\n"+
		"/dev/urandom character special file 666 1:9\n"+
		"/dev/tty character special file 666 5:0\n", out.String())
}

func TestARunFindsItsOpenFilesUnderDev(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "readlink /dev/fd /dev/stdin /dev/stdout /dev/stderr"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "/proc/self/fd\n/proc/self/fd/0\n/proc/self/fd/1\n/proc/self/fd/2\n", out.String())
}

func TestARunOpensAPseudoTerminalOfItsOwn(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "exec 3<>/dev/ptmx; ls /dev/pts"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "0\nptmx\n", out.String())
}

func TestARunSharesMemoryInDevShm(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "stat -c %a /dev/shm"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "1777\n", out.String())
}

func TestARunSeesNoneOfTheBuildersDisks(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "ls /dev"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.NotContains(t, out.String(), "vda")
}

func TestADeviceOneRunRemovesIsThereForTheNext(t *testing.T) {
	// arrange
	root := onBase(t)
	removing := protocol.Run{Command: "rm /dev/null"}
	code, err := sandbox.Run(context.Background(), root, removing, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	run := protocol.Run{Command: "test -c /dev/null"}

	// act
	code, err = sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}
