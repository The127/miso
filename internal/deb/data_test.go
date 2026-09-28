package deb_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/deb"
	"github.com/The127/miso/internal/deb/debtest"
)

func read(t *testing.T, r io.Reader) string {
	t.Helper()

	content, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(content)
}

func TestTheDataOfAPackageComesFromItsDataTarXz(t *testing.T) {
	// arrange
	pkg := debtest.Archive(t, debtest.Member{Name: "debian-binary", Content: "2.0\n"}, debtest.Member{Name: "control.tar.xz", Content: "the control"}, debtest.Member{Name: "data.tar.xz", Content: "the data"})

	// act
	data, err := deb.Data(bytes.NewReader(pkg))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "the data", read(t, data))
}

func TestAPackageWhoseDataIsPackedOtherwiseIsRefusedNamingIt(t *testing.T) {
	// arrange
	pkg := debtest.Archive(t, debtest.Member{Name: "debian-binary", Content: "2.0\n"}, debtest.Member{Name: "data.tar.zst", Content: "the data"})

	// act
	_, err := deb.Data(bytes.NewReader(pkg))

	// assert
	assert.ErrorContains(t, err, "data.tar.zst")
	assert.ErrorContains(t, err, "data.tar.xz")
}

func TestSomethingThatIsNoPackageIsRefused(t *testing.T) {
	// arrange
	notADeb := strings.NewReader("<html>404 not found</html>")

	// act
	_, err := deb.Data(notADeb)

	// assert
	assert.ErrorContains(t, err, "not a Debian package")
}
