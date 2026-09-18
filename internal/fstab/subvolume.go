package fstab

import "strings"

// RootSubvolume is the subvolume the root line of an fstab mounts, if any.
func RootSubvolume(entries []Entry) (string, bool) {
	for _, entry := range entries {
		if entry.Target != "/" {
			continue
		}

		if subvolume, found := entry.Subvolume(); found {
			return subvolume, true
		}
	}

	return "", false
}

// Subvolume is the subvolume a line mounts, named from the top level, if any.
func (e Entry) Subvolume() (string, bool) {
	for _, option := range e.Options {
		if subvolume, found := strings.CutPrefix(option, "subvol="); found {
			return strings.TrimPrefix(subvolume, "/"), true
		}
	}

	return "", false
}
