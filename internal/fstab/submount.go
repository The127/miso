package fstab

import "slices"

// Submounts are the lines that mount more of the root's own file system
// below it.
func Submounts(entries []Entry) []Entry {
	var source string

	for _, entry := range entries {
		if entry.Target == "/" {
			source = entry.Source
		}
	}

	var submounts []Entry

	for _, entry := range entries {
		if entry.Target != "/" && entry.Source == source && !slices.Contains(entry.Options, "noauto") {
			submounts = append(submounts, entry)
		}
	}

	return submounts
}
