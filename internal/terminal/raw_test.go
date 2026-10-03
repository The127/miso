package terminal_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/terminal"
)

// pty is a terminal made for a test, and answers its two ends. The slave is
// what a program takes as its terminal, and the master is the other side.
func pty(t *testing.T) (master, slave *os.File) {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|unix.O_NOCTTY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = master.Close() })

	fd := int(master.Fd())
	require.NoError(t, unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0))
	number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	require.NoError(t, err)
	slave, err = os.OpenFile("/dev/pts/"+strconv.Itoa(number), os.O_RDWR|unix.O_NOCTTY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = slave.Close() })

	return master, slave
}

func TestARawTerminalNeitherEchoesNorWaitsForALine(t *testing.T) {
	// arrange
	_, slave := pty(t)

	// act
	_, err := terminal.Raw(slave)

	// assert
	require.NoError(t, err)
	state, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
	require.NoError(t, err)
	assert.Zero(t, state.Lflag&(unix.ECHO|unix.ICANON|unix.ISIG))
}

func TestRestoringATerminalPutsItsModeBack(t *testing.T) {
	// arrange
	_, slave := pty(t)
	fd := int(slave.Fd())
	before, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	require.NoError(t, err)
	restore, err := terminal.Raw(slave)
	require.NoError(t, err)

	// act
	err = restore()

	// assert
	require.NoError(t, err)
	after, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestAFileThatIsNoTerminalCannotBeMadeRaw(t *testing.T) {
	// arrange
	file, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })

	// act
	_, err = terminal.Raw(file)

	// assert
	assert.Error(t, err)
}
