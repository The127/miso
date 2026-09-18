package fstab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/fstab"
)

func TestTheOtherLinesOfTheRootFileSystemAreSubmounts(t *testing.T) {
	// arrange
	root := fstab.Entry{Source: "UUID=15c2", Target: "/", Type: "btrfs", Options: []string{"subvol=root"}}
	home := fstab.Entry{Source: "UUID=15c2", Target: "/home", Type: "btrfs", Options: []string{"subvol=home"}}
	efi := fstab.Entry{Source: "UUID=5BCC", Target: "/boot/efi", Type: "vfat"}
	variable := fstab.Entry{Source: "UUID=15c2", Target: "/var", Type: "btrfs", Options: []string{"subvol=var"}}

	// act
	submounts := fstab.Submounts([]fstab.Entry{root, home, efi, variable})

	// assert
	assert.Equal(t, []fstab.Entry{home, variable}, submounts)
}
