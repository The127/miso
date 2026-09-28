package check_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
)

// connection is the host's end of a check's connection, whose shell has
// answered already.
type connection struct {
	answer  *strings.Reader
	written bytes.Buffer

	// what was written when the write side was closed
	closed *string
}

func answering(answer string) *connection {
	return &connection{answer: strings.NewReader(answer)}
}

func (c *connection) Read(p []byte) (int, error) { return c.answer.Read(p) }

func (c *connection) Write(p []byte) (int, error) { return c.written.Write(p) }

func (c *connection) CloseWrite() error {
	written := c.written.String()
	c.closed = &written

	return nil
}

func TestACheckIsWrittenToItsConnectionWhichIsThenClosedForWriting(t *testing.T) {
	// arrange
	conn := answering("\nmiso-exit 0\n")

	// act
	_, err := check.Run(conn, "command -v htop")

	// assert
	require.NoError(t, err)
	require.NotNil(t, conn.closed)
	assert.Equal(t, "command -v htop\n", *conn.closed)
}

func TestWhatACheckPrintedComesBackWithItsExitCode(t *testing.T) {
	// arrange
	conn := answering("hello\n\nmiso-exit 3\n")

	// act
	result, err := check.Run(conn, "echo hello; exit 3")

	// assert
	require.NoError(t, err)
	assert.Equal(t, check.Result{Output: "hello\n", Code: 3}, result)
}
