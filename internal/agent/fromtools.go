package agent

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// fromTools opens a file the tools wrote. They run the build's code, so a
// name may be a FIFO no one ever writes to, which an open must not wait on.
const fromTools = os.O_RDONLY | unix.O_NONBLOCK

// regularInfo is what an open file is when it is a regular file, and an
// error naming what it was meant to be when it is not.
func regularInfo(file *os.File, what string) (os.FileInfo, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is no regular file", what)
	}

	return info, nil
}
