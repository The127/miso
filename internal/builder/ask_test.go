package builder_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
)

const agentName = "miso 1.2.0"

func TestEveryRequestIsAskedInOrderEachOnItsOwnConnection(t *testing.T) {
	// arrange
	agent := &recording{}
	dial, dials := dialling(agent)
	requests := []protocol.Message{
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	}

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, requests, agent.asked)
	assert.Equal(t, 2, *dials)
}

func TestWhatTheAgentWritesReachesTheOutput(t *testing.T) {
	// arrange
	dial, _ := dialling(saying{"unpacking\n"})
	requests := []protocol.Message{
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	}
	var out bytes.Buffer

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "unpacking\nunpacking\n", out.String())
}

func TestTheAgentIsWaitedForUntilItListens(t *testing.T) {
	// arrange
	agent := &recording{}
	dial, _ := dialling(agent)
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(t.Context(), running{}, booting(2, dial), agentName, requests, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, requests, agent.asked)
}

func TestAVMThatStopsBeforeItsAgentListensFailsWithItsReason(t *testing.T) {
	// arrange
	reason := errors.New("QEMU stopped: exit status 1: could not open kernel")
	dial, _ := dialling(&recording{})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(t.Context(), stopped{reason}, booting(math.MaxInt, dial), agentName, requests, io.Discard)

	// assert
	assert.ErrorIs(t, err, reason)
}

func TestAVMThatStopsWithoutAReasonBeforeItsAgentListensSaysOnlyThat(t *testing.T) {
	// arrange
	dial, _ := dialling(&recording{})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(t.Context(), stopped{}, booting(math.MaxInt, dial), agentName, requests, io.Discard)

	// assert
	assert.EqualError(t, err, "the builder VM stopped before its agent listened")
}

func TestAVMThatDiesDuringAStepFailsWithItsReason(t *testing.T) {
	// arrange
	vm := &dying{done: make(chan struct{}), reason: errors.New("QEMU stopped: signal: killed")}
	requests := []protocol.Message{protocol.Run{Key: "step", Command: "true"}}

	// act
	err := builder.Ask(t.Context(), vm, vm.dial, agentName, requests, io.Discard)

	// assert
	assert.ErrorIs(t, err, vm.reason)
}

func TestAVMThatDiesDuringAStepWithoutAReasonSaysOnlyThat(t *testing.T) {
	// arrange
	vm := &dying{done: make(chan struct{})}
	requests := []protocol.Message{protocol.Run{Key: "step", Command: "true"}}

	// act
	err := builder.Ask(t.Context(), vm, vm.dial, agentName, requests, io.Discard)

	// assert
	assert.EqualError(t, err, "the builder VM stopped during a step")
}

func TestACancelledBuildCancelsTheRunningStep(t *testing.T) {
	// arrange
	ctx, cancel := context.WithCancel(t.Context())
	agent := waiting{started: make(chan struct{}), cancelled: make(chan struct{})}
	dial, _ := dialling(agent)
	requests := []protocol.Message{protocol.Run{Key: "step", Command: "sleep infinity"}}
	go func() {
		<-agent.started
		cancel()
	}()

	// act
	err := builder.Ask(ctx, running{}, dial, agentName, requests, io.Discard)

	// assert
	require.ErrorIs(t, err, context.Canceled)
	assert.Eventually(t, func() bool {
		select {
		case <-agent.cancelled:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestACancelledBuildStopsWaitingForTheAgent(t *testing.T) {
	// arrange
	ctx, cancel := context.WithCancel(t.Context())
	dial, _ := dialling(&recording{})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}
	time.AfterFunc(50*time.Millisecond, cancel)

	// act
	err := builder.Ask(ctx, running{}, booting(math.MaxInt, dial), agentName, requests, io.Discard)

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}

func TestACancelledBuildWhoseVMWasStoppedByTheCancelSaysCancelled(t *testing.T) {
	// arrange
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	dial, _ := dialling(&recording{})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}

	// act
	err := builder.Ask(ctx, stopped{errors.New("QEMU stopped: signal: killed")}, booting(math.MaxInt, dial), agentName, requests, io.Discard)

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAFailedCommandFailsTheBuildWithoutWaitingForTheVM(t *testing.T) {
	// arrange
	dial, _ := dialling(exiting{code: 1})
	requests := []protocol.Message{protocol.Run{Key: "step", Command: "false"}}
	start := time.Now()

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, io.Discard)

	// assert
	require.ErrorIs(t, err, protocol.ErrCommandFailed)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestAnAgentThatFailsFailsTheBuildWithoutWaitingForTheVM(t *testing.T) {
	// arrange
	dial, _ := dialling(failing{errors.New("no space left on device")})
	requests := []protocol.Message{protocol.Import{Key: "base", Digest: "sha256:aaaa"}}
	start := time.Now()

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, io.Discard)

	// assert
	require.ErrorIs(t, err, protocol.ErrAgentFailed)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

// running is a VM that keeps running.
type running struct{}

func (running) Done() <-chan struct{} { return nil }

func (running) Err() error { return nil }

// stopped is a VM that has stopped for a reason.
type stopped struct {
	reason error
}

func (s stopped) Done() <-chan struct{} {
	done := make(chan struct{})
	close(done)

	return done
}

func (s stopped) Err() error { return s.reason }

// dying is a VM that dies while its agent works on a request. Its
// connection breaks first and the VM counts as stopped a moment later, as
// when QEMU is killed.
type dying struct {
	done   chan struct{}
	reason error
}

func (d *dying) Done() <-chan struct{} { return d.done }

func (d *dying) Err() error { return d.reason }

func (d *dying) dial() (io.ReadWriteCloser, error) {
	host, agent := net.Pipe()
	go func() {
		_, _ = bufio.NewReader(agent).ReadString('\n')
		_ = agent.Close()
		time.Sleep(50 * time.Millisecond)
		close(d.done)
	}()

	return host, nil
}

// booting fails the first dials, the way dialling a VM fails before its
// agent listens, and dials after that.
func booting(fails int, dial func() (io.ReadWriteCloser, error)) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		if fails > 0 {
			fails--

			return nil, errors.New("connection reset by peer")
		}

		return dial()
	}
}

// dialling hands out connections to an agent that does its work with the
// runner, one request per connection as the real one answers, and counts
// them.
func dialling(runner protocol.Runner) (func() (io.ReadWriteCloser, error), *int) {
	dials := 0
	dial := func() (io.ReadWriteCloser, error) {
		dials++
		host, agent := net.Pipe()
		go func() {
			_ = protocol.New(agentName, agent, agent).Serve(runner)
			_ = agent.Close()
		}()

		return host, nil
	}

	return dial, &dials
}

// recording is an agent whose work succeeds and who notes what it was asked.
type recording struct {
	mu    sync.Mutex
	asked []protocol.Message
}

func (r *recording) Run(_ context.Context, run protocol.Run, _ io.Writer) (int, error) {
	r.note(run)

	return 0, nil
}

func (r *recording) Import(_ context.Context, request protocol.Import, _ io.Writer) error {
	r.note(request)

	return nil
}

func (r *recording) note(request protocol.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.asked = append(r.asked, request)
}

// saying is an agent whose work succeeds and writes the same words each
// time.
type saying struct {
	words string
}

func (s saying) Run(_ context.Context, _ protocol.Run, out io.Writer) (int, error) {
	_, err := io.WriteString(out, s.words)

	return 0, err
}

func (s saying) Import(_ context.Context, _ protocol.Import, out io.Writer) error {
	_, err := io.WriteString(out, s.words)

	return err
}

// exiting is an agent whose commands exit with a code.
type exiting struct {
	code int
}

func (e exiting) Run(context.Context, protocol.Run, io.Writer) (int, error) {
	return e.code, nil
}

func (e exiting) Import(context.Context, protocol.Import, io.Writer) error {
	return nil
}

// failing is an agent that fails at its own work.
type failing struct {
	err error
}

func (f failing) Run(context.Context, protocol.Run, io.Writer) (int, error) {
	return 0, f.err
}

func (f failing) Import(context.Context, protocol.Import, io.Writer) error {
	return f.err
}

// waiting is an agent whose run lasts until it is cancelled.
type waiting struct {
	started   chan struct{}
	cancelled chan struct{}
}

func (w waiting) Run(ctx context.Context, _ protocol.Run, _ io.Writer) (int, error) {
	close(w.started)
	<-ctx.Done()
	close(w.cancelled)

	return 0, ctx.Err()
}

func (w waiting) Import(context.Context, protocol.Import, io.Writer) error {
	return nil
}
