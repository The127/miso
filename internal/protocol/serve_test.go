package protocol_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

type runner struct {
	writes string
	code   int
	err    error
}

func (r runner) Run(_ protocol.Run, out io.Writer) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.writes == "" {
		return r.code, nil
	}

	_, err := io.WriteString(out, r.writes)

	return r.code, err
}

func TestARunThatWorksIsDone(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "echo hello"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{writes: "hello\n"})

	// assert
	require.NoError(t, err)
	output, err := host.Receive()
	require.NoError(t, err)
	done, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Output{Bytes: []byte("hello\n")}, output)
	assert.Equal(t, protocol.Done{}, done)
}

func TestARunThatExitsNonZeroIsExited(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "apt-get install nope"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{code: 100})

	// assert
	require.NoError(t, err)
	exited, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Exited{Code: 100}, exited)
}

func TestARunnerErrorIsFailed(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "apt-get update"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{err: errors.New("mount overlay: no space left on device")})

	// assert
	require.NoError(t, err)
	failed, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Failed{Reason: "mount overlay: no space left on device"}, failed)
}

func TestAMessageThatIsNoRequestIsRefused(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Done{}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{})

	// assert
	require.NoError(t, err)
	failed, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Failed{Reason: "protocol.Done is not a request"}, failed)
}

type watched struct {
	ran bool
}

func (w *watched) Run(protocol.Run, io.Writer) (int, error) {
	w.ran = true

	return 0, nil
}

func TestARequestFromAnotherAgentIsRefusedBeforeItRuns(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "apt-get update"}))
	agent := protocol.New("miso 1.3.0", &requests, &replies)
	runner := &watched{}

	// act
	err := agent.Serve(runner)

	// assert
	require.NoError(t, err)
	_, err = host.Receive()
	assert.ErrorIs(t, err, protocol.ErrAnotherAgent)
	assert.False(t, runner.ran)
}
