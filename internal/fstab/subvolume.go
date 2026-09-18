package fstab

import "strings"

// RootSubvolume is the subvolume the root line of an fstab mounts, if any.
func RootSubvolume(entries []Entry) (string, bool) {
	for _, entry := range entries {
		if entry.Target != "/" {
			continue
		}

		for _, option := range entry.Options {
			if subvolume, found := strings.CutPrefix(option, "subvol="); found {
				return strings.TrimPrefix(subvolume, "/"), true
			}
		}
	}

	return "", false
}
