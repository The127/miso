package builderkernel

import (
	"bufio"
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
