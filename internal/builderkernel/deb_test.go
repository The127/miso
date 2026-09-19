package builderkernel_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
)

// member is a file in an ar archive, as a package holds them.
type member struct {
	name    string
	content string
}

// archive is an ar archive of the members, the shape of a Debian package.
func archive(t *testing.T, members ...member) []byte {
	t.Helper()

	var archive bytes.Buffer
	archive.WriteString("!<arch>\n")
	for _, file := range members {
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", file.name, 0, 0, 0, "100644", len(file.content))
		require.Len(t, header, 60)
		archive.WriteString(header)
		archive.WriteString(file.content)
		if len(file.content)%2 == 1 {
			archive.WriteString("\n")
		}
	}

	return archive.Bytes()
}

func read(t *testing.T, r io.Reader) string {
	t.Helper()

	content, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(content)
}

func TestTheDataOfAPackageComesFromItsDataTarXz(t *testing.T) {
	// arrange
	deb := archive(t, member{"debian-binary", "2.0\n"}, member{"control.tar.xz", "the control"}, member{"data.tar.xz", "the data"})

	// act
	data, err := builderkernel.Data(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "the data", read(t, data))
}

func TestAPackageWhoseDataIsPackedOtherwiseIsRefusedNamingIt(t *testing.T) {
	// arrange
	deb := archive(t, member{"debian-binary", "2.0\n"}, member{"data.tar.zst", "the data"})

	// act
	_, err := builderkernel.Data(bytes.NewReader(deb))

	// assert
	assert.ErrorContains(t, err, "data.tar.zst")
	assert.ErrorContains(t, err, "data.tar.xz")
}
