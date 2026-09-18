package fstab

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrBadLine is a line of an fstab that names no file system to mount.
var ErrBadLine = errors.New("bad fstab line")

// Entry is one line of an fstab.
type Entry struct {
	Source  string
	Target  string
	Type    string
	Options []string
}

// Parse reads the entries of an fstab.
func Parse(r io.Reader) ([]Entry, error) {
	var entries []Entry

	lines := bufio.NewScanner(r)
	for number := 1; lines.Scan(); number++ {
		fields := strings.Fields(lines.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		if len(fields) < 3 {
			return nil, fmt.Errorf("line %d: %w", number, ErrBadLine)
		}

		entry := Entry{Source: fields[0], Target: fields[1], Type: fields[2]}
		if len(fields) > 3 {
			entry.Options = strings.Split(fields[3], ",")
		}

		entries = append(entries, entry)
	}

	return entries, lines.Err()
}
