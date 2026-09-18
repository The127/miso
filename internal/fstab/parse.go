package fstab

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
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

		entry := Entry{Source: unescape(fields[0]), Target: unescape(fields[1]), Type: fields[2]}
		if len(fields) > 3 {
			entry.Options = strings.Split(fields[3], ",")
		}

		entries = append(entries, entry)
	}

	return entries, lines.Err()
}

// unescape turns each backslash with three octal digits into the byte they
// stand for, which is how an fstab writes a space in a name.
func unescape(name string) string {
	var out strings.Builder

	for i := 0; i < len(name); i++ {
		if name[i] == '\\' && i+4 <= len(name) {
			if b, err := strconv.ParseUint(name[i+1:i+4], 8, 8); err == nil {
				out.WriteByte(byte(b))
				i += 3

				continue
			}
		}

		out.WriteByte(name[i])
	}

	return out.String()
}
