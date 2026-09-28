package protocol_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestAMessageThatIsNoEntryIsAnErrorAndNotTheEnd(t *testing.T) {
	// arrange
	var requests bytes.Buffer
	host := protocol.New("miso 1.2.0", bytes.NewReader(nil), &requests)
	require.NoError(t, host.Send(protocol.Done{}))
	agent := protocol.New("miso 1.2.0", &requests, io.Discard)

	// act
	_, _, err := protocol.EntriesOf(agent).Next()

	// assert
	assert.NotErrorIs(t, err, io.EOF)
	assert.EqualError(t, err, "protocol.Done is not an entry")
}
