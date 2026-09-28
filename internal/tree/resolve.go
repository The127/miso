package tree

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"golang.org/x/sys/unix"
)

// openat2 is the kernel's, and only a test answers for it.
var openat2 = unix.Openat2

// sourceRoot opens the directory a source lives in and gives the source's
// name there. A link on the way resolves inside the stage, even one that
// names an absolute path, since the stage is its own root file system. The
// source itself is never followed.
func sourceRoot(stage int, source string) (*os.Root, string, error) {
	dir, name := path.Split(path.Clean("/" + source))
	if name == "" {
		name = "."
	}

	fd, err := inStage(stage, dir)
	if err != nil {
		return nil, "", &fs.PathError{Op: "open", Path: source, Err: err}
	}

	defer func() { _ = unix.Close(fd) }()

	root, err := os.OpenRoot(fmt.Sprintf("/proc/self/fd/%d", fd))
	if err != nil {
		return nil, "", &fs.PathError{Op: "open", Path: source, Err: err}
	}

	return root, name, nil
}

// races is how often inStage tries again after a rename raced it, far more
// than a builder VM that does little else besides a build ever needs.
const races = 64

// inStage opens a directory of the stage, its links resolved inside the
// stage. A lookup through .. that a rename or mount anywhere raced fails
// with EAGAIN, and the kernel asks to try again.
func inStage(stage int, dir string) (int, error) {
	how := &unix.OpenHow{
		Flags:   unix.O_PATH | unix.O_DIRECTORY | unix.O_CLOEXEC,
		Resolve: unix.RESOLVE_IN_ROOT | unix.RESOLVE_NO_MAGICLINKS,
	}

	for range races {
		fd, err := openat2(stage, dir, how)
		if !errors.Is(err, unix.EAGAIN) {
			return fd, err
		}
	}

	return openat2(stage, dir, how)
}
