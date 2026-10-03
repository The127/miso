package terminal

import (
	"os"

	"golang.org/x/sys/unix"
)

// Raw gives the terminal no line editing, echo, signals or output
// processing of its own, so every key reaches whoever reads it and the
// machine on the other end does the rest, and answers what puts it back.
func Raw(f *os.File) (restore func() error, err error) {
	fd := int(f.Fd())
	before, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	raw := *before
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &raw); err != nil {
		return nil, err
	}

	return func() error { return unix.IoctlSetTermios(fd, unix.TCSETS, before) }, nil
}
