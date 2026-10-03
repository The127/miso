package terminal

import (
	"os"

	"golang.org/x/sys/unix"
)

// Size is how many rows and columns the terminal has.
func Size(f *os.File) (rows, cols uint16, err error) {
	size, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}

	return size.Row, size.Col, nil
}
