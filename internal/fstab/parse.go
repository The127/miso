package fstab

import (
	"bufio"
	"io"
	"strings"
)

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
	for lines.Scan() {
		fields := strings.Fields(lines.Text())
		entries = append(entries, Entry{
			Source:  fields[0],
			Target:  fields[1],
			Type:    fields[2],
			Options: strings.Split(fields[3], ","),
		})
	}

	return entries, lines.Err()
}
