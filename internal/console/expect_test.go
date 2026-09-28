package console_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/console"
)

func TestExpectReturnsWhatCameBeforeTheMatch(t *testing.T) {
	// arrange
	serial := strings.NewReader("booting\nlogin: ")

	// act
	before, err := console.Expect(serial, regexp.MustCompile(`login: `))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "booting\n", before)
}
