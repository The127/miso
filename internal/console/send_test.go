package console_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/console"
)

func TestSendTypesTheTextOnTheConsole(t *testing.T) {
	// arrange
	var keyboard bytes.Buffer
	tty := console.New(strings.NewReader(""), &keyboard)

	// act
	err := tty.Send("root\n")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "root\n", keyboard.String())
}
