//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestAShellRunsOnATerminal(t *testing.T) {
	// arrange
	root := onBase(t)
	// the terminal echoes what is typed, so only the output has the word whole
	in := strings.NewReader("test -t 0 && echo ter\"\"minal\nexit\n")
	var out bytes.Buffer

	// act
	code, err := sandbox.Shell(context.Background(), root, t.TempDir(), protocol.Shell{}, in, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "terminal\r\n")
}
