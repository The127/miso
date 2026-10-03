package terminal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/terminal"
)

func TestTheSizeOfATerminalIsItsRowsAndColumns(t *testing.T) {
	// arrange
	master, slave := pty(t)
	require.NoError(t, unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 24, Col: 80}))

	// act
	rows, cols, err := terminal.Size(slave)

	// assert
	require.NoError(t, err)
	assert.Equal(t, uint16(24), rows)
	assert.Equal(t, uint16(80), cols)
}
