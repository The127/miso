package buildcontext_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
)

func TestPackingAFileHandsItsEntryAndItsContent(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "motd", "hello\n")
	chmod(t, dir, "motd", 0o644)
	var entries []buildcontext.Entry
	var contents []string

	// act
	err := open(t, dir).Pack("motd", func(entry buildcontext.Entry, content io.Reader) error {
		read, err := io.ReadAll(content)
		entries = append(entries, entry)
		contents = append(contents, string(read))

		return err
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []buildcontext.Entry{{Kind: "file", Path: ".", Mode: 0o644, Size: 6}}, entries)
	assert.Equal(t, []string{"hello\n"}, contents)
}

func TestPackingATreeHandsItsEntriesInTheOrderOfItsDigest(t *testing.T) {
	// arrange
	dir := t.TempDir()
	mkdir(t, dir, "etc")
	chmod(t, dir, "etc", 0o755)
	write(t, dir, "etc/motd", "hello\n")
	chmod(t, dir, "etc/motd", 0o644)
	symlink(t, dir, "etc/issue", "motd")
	var entries []buildcontext.Entry
	var contents []string

	// act
	err := open(t, dir).Pack("etc", func(entry buildcontext.Entry, content io.Reader) error {
		read, err := io.ReadAll(content)
		entries = append(entries, entry)
		contents = append(contents, string(read))

		return err
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []buildcontext.Entry{
		{Kind: "directory", Path: ".", Mode: 0o755},
		{Kind: "link", Path: "issue", Target: "motd"},
		{Kind: "file", Path: "motd", Mode: 0o644, Size: 6},
	}, entries)
	assert.Equal(t, []string{"", "", "hello\n"}, contents)
}
