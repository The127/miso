package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"golang.org/x/sys/unix"
)

// sourceRoot opens the directory a source lives in and gives the source's
// name there. A link on the way resolves inside the stage, even one that
// names an absolute path, since the stage is its own root file system. The
// source itself is never followed.
func sourceRoot(stage int, source string) (*os.Root, string, error) {
	dir, name := path.Split(path.Clean("/" + source))
	if name == "" {
		name = "."
	}

	fd, err := unix.Openat2(stage, dir, &unix.OpenHow{
		Flags:   unix.O_PATH | unix.O_DIRECTORY | unix.O_CLOEXEC,
		Resolve: unix.RESOLVE_IN_ROOT | unix.RESOLVE_NO_MAGICLINKS,
	})
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
