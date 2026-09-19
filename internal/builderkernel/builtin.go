package builderkernel

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"strings"
)

// builtinFile is what a package calls the list of what its kernel has built in
const builtinFile = "modules.builtin"

// names are the modules a modules.builtin lists, where every line is the
// path the module would have been built to.
func names(file io.Reader) ([]string, error) {
	var found []string

	lines := bufio.NewScanner(file)
	for lines.Scan() {
		found = append(found, strings.TrimSuffix(path.Base(lines.Text()), moduleExt))
	}

	return found, lines.Err()
}

// builtin names everything the kernel of a package has built in. None of it
// is a module to load.
func builtin(deb io.Reader) ([]string, error) {
	files, failed := eachFile(deb)
	for name, file := range files {
		if path.Base(name) != builtinFile {
			continue
		}

		return names(file)
	}

	if err := failed(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("the package holds no %s", builtinFile)
}
