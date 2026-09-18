package fstab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/fstab"
)

func TestTheRootSubvolumeIsNamedByTheRootLine(t *testing.T) {
	// arrange
	entries := []fstab.Entry{
		{Source: "UUID=15c2", Target: "/var", Type: "btrfs", Options: []string{"subvol=var"}},
		{Source: "UUID=15c2", Target: "/", Type: "btrfs", Options: []string{"compress=zstd:1", "subvol=root"}},
	}

	// act
	subvolume, found := fstab.RootSubvolume(entries)

	// assert
	assert.True(t, found)
	assert.Equal(t, "root", subvolume)
}

func TestARootSubvolumeIsNamedFromTheTopLevel(t *testing.T) {
	// arrange
	entries := []fstab.Entry{
		{Source: "UUID=15c2", Target: "/", Type: "btrfs", Options: []string{"subvol=/@/.snapshots/1/snapshot"}},
	}

	// act
	subvolume, found := fstab.RootSubvolume(entries)

	// assert
	assert.True(t, found)
	assert.Equal(t, "@/.snapshots/1/snapshot", subvolume)
}
