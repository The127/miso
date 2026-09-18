package fstab_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/fstab"
)

func TestALineIsAnEntry(t *testing.T) {
	// arrange
	text := "UUID=15c2 / btrfs compress=zstd:1,defaults,subvol=root 0 1\n"

	// act
	entries, err := fstab.Parse(strings.NewReader(text))

	// assert
	require.NoError(t, err)
	assert.Equal(t, []fstab.Entry{{
		Source:  "UUID=15c2",
		Target:  "/",
		Type:    "btrfs",
		Options: []string{"compress=zstd:1", "defaults", "subvol=root"},
	}}, entries)
}

func TestALineWithoutATypeIsABadLine(t *testing.T) {
	// arrange
	text := "# fstab\nUUID=15c2 /\n"

	// act
	_, err := fstab.Parse(strings.NewReader(text))

	// assert
	assert.ErrorIs(t, err, fstab.ErrBadLine)
	assert.ErrorContains(t, err, "line 2")
}

func TestCommentsAndBlankLinesAreNoEntries(t *testing.T) {
	// arrange
	text := "# /etc/fstab\n\n   \n  # indented\nUUID=15c2 / btrfs defaults 0 1\n"

	// act
	entries, err := fstab.Parse(strings.NewReader(text))

	// assert
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "/", entries[0].Target)
}
