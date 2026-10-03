package protocol_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestWhatTheUserTypesInAShellReachesTheAgent(t *testing.T) {
	// arrange
	replies, agentOut := io.Pipe()
	requests, hostOut := io.Pipe()
	host := protocol.New("miso 1.2.0", replies, hostOut)
	agent := protocol.New("miso 1.2.0", requests, agentOut)
	typed := make(chan protocol.Message, 2)
	go func() {
		for range 2 {
			message, err := agent.Receive()
			if err != nil {
				return
			}

			typed <- message
		}

		_ = agent.Send(protocol.Done{})
	}()

	// act
	_, err := host.AskShell(protocol.Shell{}, protocol.Terminal{In: strings.NewReader("ls\n")}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, protocol.Shell{}, <-typed)
	assert.Equal(t, protocol.Input{Bytes: []byte("ls\n")}, <-typed)
}

func TestAShellThatExitsWithACodeAnswersTheCode(t *testing.T) {
	// arrange
	var replies bytes.Buffer
	agent := protocol.New("miso 1.2.0", bytes.NewReader(nil), &replies)
	require.NoError(t, agent.Send(protocol.Exited{Code: 3}))
	host := protocol.New("miso 1.2.0", &replies, io.Discard)

	// act
	code, err := host.AskShell(protocol.Shell{}, protocol.Terminal{In: strings.NewReader("")}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestWhatAShellPrintsReachesTheWriter(t *testing.T) {
	// arrange
	var replies bytes.Buffer
	agent := protocol.New("miso 1.2.0", bytes.NewReader(nil), &replies)
	require.NoError(t, agent.Send(protocol.Output{Bytes: []byte("# ")}))
	require.NoError(t, agent.Send(protocol.Done{}))
	host := protocol.New("miso 1.2.0", &replies, io.Discard)
	var out bytes.Buffer

	// act
	_, err := host.AskShell(protocol.Shell{}, protocol.Terminal{In: strings.NewReader("")}, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "# ", out.String())
}

func TestTheSizeTheTerminalOfTheUserChangesToReachesTheAgent(t *testing.T) {
	// arrange
	replies, agentOut := io.Pipe()
	requests, hostOut := io.Pipe()
	host := protocol.New("miso 1.2.0", replies, hostOut)
	agent := protocol.New("miso 1.2.0", requests, agentOut)
	resized := make(chan protocol.Resize, 1)
	resized <- protocol.Resize{Rows: 30, Cols: 100}
	got := make(chan protocol.Message, 2)
	go func() {
		for range 2 {
			message, err := agent.Receive()
			if err != nil {
				return
			}

			got <- message
		}

		_ = agent.Send(protocol.Done{})
	}()
	// a shell that never ends the input, as a terminal does not
	in, _ := io.Pipe()

	// act
	_, err := host.AskShell(protocol.Shell{}, protocol.Terminal{In: in, Resized: resized}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, protocol.Shell{}, <-got)
	assert.Equal(t, protocol.Resize{Rows: 30, Cols: 100}, <-got)
}
