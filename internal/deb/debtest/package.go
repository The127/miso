package debtest

import (
	"archive/tar"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ulikunitz/xz"
)

// Member is a file in an ar archive, as a package holds them.
type Member struct {
	Name    string
	Content string
}

// Archive is an ar archive of the members, the shape of a Debian package.
func Archive(t *testing.T, members ...Member) []byte {
	t.Helper()

	var archive bytes.Buffer
	archive.WriteString("!<arch>\n")
	for _, file := range members {
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", file.Name, 0, 0, 0, "100644", len(file.Content))
		require.Len(t, header, 60)
		archive.WriteString(header)
		archive.WriteString(file.Content)
		if len(file.Content)%2 == 1 {
			archive.WriteString("\n")
		}
	}

	return archive.Bytes()
}

// Package is a Debian package whose data holds the files.
func Package(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var packed bytes.Buffer

	compressed, err := xz.NewWriter(&packed)
	require.NoError(t, err)

	writer := tar.NewWriter(compressed)
	for name, content := range files {
		require.NoError(t, writer.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}))
		_, err := writer.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())
	require.NoError(t, compressed.Close())

	return Archive(t, Member{"debian-binary", "2.0\n"}, Member{"data.tar.xz", packed.String()})
}
