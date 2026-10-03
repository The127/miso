package builder

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
)

// sumsFile is the name of the file that lists the sums of the files of a
// directory, which systemd-sysupdate reads.
const sumsFile = "SHA256SUMS"

// summed is the file of a disk whose sum is listed once it is whole.
type summed struct {
	plain
}

func (s summed) Close() error {
	if err := s.plain.Close(); err != nil {
		return err
	}

	return listSum(s.dir, s.name)
}

// listSum puts the sum of the file a name gives into the SHA256SUMS of its
// directory, whose lines stay in the order of their names.
func listSum(dir, name string) error {
	return inRoot(dir, func(root *os.Root) error {
		sum, err := sumOf(root, name)
		if err != nil {
			return err
		}

		listing := path.Join(path.Dir(name), sumsFile)

		lines, err := linesOf(root, listing)
		if err != nil {
			return err
		}

		lines[path.Base(name)] = sum

		var text strings.Builder
		for _, file := range slices.Sorted(maps.Keys(lines)) {
			text.WriteString(lines[file] + "  " + file + "\n")
		}

		return root.WriteFile(listing, []byte(text.String()), 0o644)
	})
}

// sumOf is the SHA-256 of a file, in hex.
func sumOf(root *os.Root, name string) (string, error) {
	file, err := root.Open(name)
	if err != nil {
		return "", err
	}

	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// linesOf are the sums a listing holds by the names of their files, none
// when there is no listing yet.
func linesOf(root *os.Root, listing string) (map[string]string, error) {
	lines := map[string]string{}

	file, err := root.Open(listing)
	if errors.Is(err, fs.ErrNotExist) {
		return lines, nil
	}

	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sum, name, found := strings.Cut(scanner.Text(), "  ")
		if found {
			lines[name] = sum
		}
	}

	return lines, scanner.Err()
}
