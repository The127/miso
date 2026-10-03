package builder_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
)

func TestAShellIsAskedAfterTheStepsBeforeIt(t *testing.T) {
	// arrange
	agent := &recording{}
	requests := requested(
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
		protocol.Shell{Layers: []string{"base", "step"}},
	)

	// act
	code, err := builder.Shell(t.Context(), running{}, dialling(agent), agentName, requests, nil, strings.NewReader(""), io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, messagesOf(requests), agent.asked)
}

// echoing is an agent whose shell prints what the user types.
type echoing struct {
	*recording
}

func (e echoing) Shell(_ context.Context, request protocol.Shell, in io.Reader, out io.Writer) (int, error) {
	e.note(request)

	typed := make([]byte, len("ls\n"))
	if _, err := io.ReadFull(in, typed); err != nil {
		return 0, err
	}

	_, err := out.Write(typed)

	return 0, err
}

func TestWhatTheUserTypesInAShellComesBackAsItsOutput(t *testing.T) {
	// arrange
	agent := echoing{&recording{}}
	requests := requested(protocol.Shell{Layers: []string{"base"}})
	var out bytes.Buffer

	// act
	_, err := builder.Shell(t.Context(), running{}, dialling(agent), agentName, requests, nil, strings.NewReader("ls\n"), &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "ls\n", out.String())
}

func TestRequestsWithoutAShellAtTheEndAreRefused(t *testing.T) {
	// arrange
	agent := &recording{}
	requests := requested(protocol.Import{Key: "base", Digest: "sha256:aaaa"})

	// act
	_, err := builder.Shell(t.Context(), running{}, dialling(agent), agentName, requests, nil, strings.NewReader(""), io.Discard)

	// assert
	assert.ErrorIs(t, err, builder.ErrNoShell)
	assert.Empty(t, agent.asked)
}

func TestACancelledShellEndsTheShell(t *testing.T) {
	// arrange
	ctx, cancel := context.WithCancel(t.Context())
	agent := waiting{started: make(chan struct{}), cancelled: make(chan struct{})}
	requests := requested(protocol.Shell{Layers: []string{"base"}})
	go func() {
		<-agent.started
		cancel()
	}()

	// act
	_, err := builder.Shell(ctx, running{}, dialling(agent), agentName, requests, nil, strings.NewReader(""), io.Discard)

	// assert
	require.ErrorIs(t, err, context.Canceled)
	select {
	case <-agent.cancelled:
	case <-time.After(time.Second):
		t.Fatal("the agent's shell was not cancelled")
	}
}

func TestAShellThatFailsNamesItsLineOfTheBuildFile(t *testing.T) {
	// arrange
	agent := failing{err: errors.New("no shell in the image")}
	requests := []build.Request{{Line: 3, Written: "shell before RUN make", Message: protocol.Shell{Layers: []string{"base"}}}}

	// act
	_, err := builder.Shell(t.Context(), running{}, dialling(agent), agentName, requests, nil, strings.NewReader(""), io.Discard)

	// assert
	assert.ErrorContains(t, err, "line 3: shell before RUN make")
	assert.ErrorContains(t, err, "no shell in the image")
}
