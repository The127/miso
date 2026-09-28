package check_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
)

// connection is the host's end of a check's connection.
type connection struct {
	written bytes.Buffer

	// what was written when the write side was closed
	closed *string
}

func (c *connection) Write(p []byte) (int, error) { return c.written.Write(p) }

func (c *connection) CloseWrite() error {
	written := c.written.String()
	c.closed = &written

	return nil
}

func TestACheckIsWrittenToItsConnectionWhichIsThenClosedForWriting(t *testing.T) {
	// arrange
	conn := &connection{}

	// act
	err := check.Run(conn, "command -v htop")

	// assert
	require.NoError(t, err)
	require.NotNil(t, conn.closed)
	assert.Equal(t, "command -v htop\n", *conn.closed)
}
