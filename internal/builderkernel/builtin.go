package builderkernel

import (
	"bufio"
	"errors"
	"io"
	"path"
	"strings"
)

// builtinFile is what a package calls the list of what its kernel has built in
const builtinFile = "modules.builtin"

// builtin names everything the kernel of a package has built in. None of it
// is a module to load.
func builtin(deb io.Reader) ([]string, error) {
	files, err := unpacked(deb)
	if err != nil {
		return nil, err
	}

	for {
		header, err := files.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		if path.Base(header.Name) != builtinFile {
			continue
		}

		var found []string

		// every line is the path a module would have been built to
		lines := bufio.NewScanner(files)
		for lines.Scan() {
			found = append(found, strings.TrimSuffix(path.Base(lines.Text()), moduleExt))
		}

		return found, lines.Err()
	}

	return nil, nil
}
