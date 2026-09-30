package builder_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
)

const agentName = "miso 1.2.0"

func TestEveryRequestIsAskedInOrderEachOnItsOwnConnection(t *testing.T) {
	// arrange
	agent := &recording{}
	dial, dials := counted(dialling(agent))
	requests := requested(
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	)

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, nil, nil, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, messagesOf(requests), agent.asked)
	assert.Equal(t, 2, *dials)
}

func TestACopyGetsItsFilesFromTheBuildContext(t *testing.T) {
	// arrange
	agent := &receiving{}
	request := protocol.Copy{Key: "step", Layers: []string{"base"}, Sources: []string{"motd"}, Digests: []string{"sha256:aaaa"}, Destination: "/etc/motd"}
	var asked []protocol.Copy
	files := func(copying protocol.Copy) protocol.Files {
		asked = append(asked, copying)

		return func(send func(protocol.Entry, io.Reader) error) error {
			return send(protocol.Entry{Kind: "file", Path: ".", Mode: 0o644, Size: 6}, strings.NewReader("hello\n"))
		}
	}

	// act
	err := builder.Ask(t.Context(), running{}, dialling(agent), agentName, requested(request), files, nil, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []protocol.Copy{request}, asked)
	assert.Equal(t, []string{"hello\n"}, agent.contents)
}

func TestWhatTheAgentWritesReachesTheOutput(t *testing.T) {
	// arrange
	dial := dialling(saying{"unpacking\n"})
	requests := requested(
		protocol.Import{Key: "base", Digest: "sha256:aaaa"},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	)
	var out bytes.Buffer

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, nil, nil, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "unpacking\nunpacking\n", out.String())
}

func TestTheAgentIsWaitedForUntilItListens(t *testing.T) {
	// arrange
	agent := &recording{}
	dial := dialling(agent)
	requests := oneImport()

	// act
	err := builder.Ask(t.Context(), running{}, booting(2, dial), agentName, requests, nil, nil, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, messagesOf(requests), agent.asked)
}

func TestAVMThatStopsBeforeItsAgentListensFailsWithItsReason(t *testing.T) {
	// arrange
	reason := errors.New("QEMU stopped: exit status 1: could not open kernel")
	requests := oneImport()

	// act
	err := builder.Ask(t.Context(), stopped{reason}, unreachable, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.ErrorIs(t, err, reason)
}

func TestAVMThatStopsWithoutAReasonBeforeItsAgentListensSaysOnlyThat(t *testing.T) {
	// arrange
	requests := oneImport()

	// act
	err := builder.Ask(t.Context(), stopped{}, unreachable, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.EqualError(t, err, "the builder VM stopped before its agent listened")
}

func TestAVMThatDiesDuringAStepFailsWithItsReason(t *testing.T) {
	// arrange
	vm := &dying{done: make(chan struct{}), reason: errors.New("QEMU stopped: signal: killed")}
	requests := requested(protocol.Run{Key: "step", Command: "true"})

	// act
	err := builder.Ask(t.Context(), vm, vm.dial, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.ErrorIs(t, err, vm.reason)
}

func TestAVMThatDiesDuringAStepWithoutAReasonSaysOnlyThat(t *testing.T) {
	// arrange
	vm := &dying{done: make(chan struct{})}
	requests := []build.Request{{Line: 2, Written: "RUN true", Message: protocol.Run{Key: "step", Command: "true"}}}

	// act
	err := builder.Ask(t.Context(), vm, vm.dial, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.EqualError(t, err, "line 2: RUN true: the builder VM stopped during a step")
}

func TestACancelledBuildCancelsTheRunningStep(t *testing.T) {
	// arrange
	ctx, cancel := context.WithCancel(t.Context())
	agent := waiting{started: make(chan struct{}), cancelled: make(chan struct{})}
	dial := dialling(agent)
	requests := requested(protocol.Run{Key: "step", Command: "sleep infinity"})
	cancelled := make(chan time.Time, 1)
	go func() {
		<-agent.started
		cancelled <- time.Now()
		cancel()
	}()

	// act
	err := builder.Ask(ctx, running{}, dial, agentName, requests, nil, nil, io.Discard)

	// assert
	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(<-cancelled), 500*time.Millisecond)
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
	requests := oneImport()
	time.AfterFunc(50*time.Millisecond, cancel)

	// act
	err := builder.Ask(ctx, running{}, unreachable, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}

func TestACancelledBuildWhoseVMWasStoppedByTheCancelSaysCancelled(t *testing.T) {
	// arrange
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	requests := oneImport()

	// act
	err := builder.Ask(ctx, stopped{errors.New("QEMU stopped: signal: killed")}, unreachable, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAnAgentThatDoesNotListenInTimeFailsTheBuild(t *testing.T) {
	// arrange
	builder.Patience(t, 100*time.Millisecond)
	requests := oneImport()

	// act
	err := builder.Ask(t.Context(), running{}, unreachable, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.ErrorContains(t, err, "the agent did not listen within 100ms")
}

func TestAnAgentOnAVMWithoutKVMIsWaitedForTenTimesAsLong(t *testing.T) {
	// arrange
	builder.Patience(t, 20*time.Millisecond)
	requests := oneImport()

	// act
	err := builder.Ask(t.Context(), onTCG{}, unreachable, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.ErrorContains(t, err, "the agent did not listen within 200ms")
}

func TestAFailedCommandFailsTheBuildWithoutWaitingForTheVM(t *testing.T) {
	// arrange
	dial := dialling(exiting{code: 1})
	requests := requested(protocol.Run{Key: "step", Command: "false"})
	start := time.Now()

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, nil, nil, io.Discard)

	// assert
	require.ErrorIs(t, err, protocol.ErrCommandFailed)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestAFailedStepIsNamedAfterItsLineOfTheBuildFile(t *testing.T) {
	// arrange
	dial := dialling(exiting{code: 1})
	requests := []build.Request{{Line: 2, Written: "RUN false", Message: protocol.Run{Key: "step", Command: "false"}}}

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, nil, nil, io.Discard)

	// assert
	assert.EqualError(t, err, "line 2: RUN false: command failed: exit code 1")
}

func TestAnAgentThatFailsFailsTheBuildWithoutWaitingForTheVM(t *testing.T) {
	// arrange
	dial := dialling(failing{errors.New("no space left on device")})
	requests := oneImport()
	start := time.Now()

	// act
	err := builder.Ask(t.Context(), running{}, dial, agentName, requests, nil, nil, io.Discard)

	// assert
	require.ErrorIs(t, err, protocol.ErrAgentFailed)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestAFileOfTheHostThatFailsFailsTheBuildWithoutWaitingForTheVM(t *testing.T) {
	// arrange
	request := protocol.Copy{Key: "step", Sources: []string{"motd"}, Digests: []string{"sha256:aaaa"}, Destination: "/etc/motd"}
	files := func(protocol.Copy) protocol.Files {
		return func(func(protocol.Entry, io.Reader) error) error {
			return errors.New("motd: changed since it was planned")
		}
	}
	start := time.Now()

	// act
	err := builder.Ask(t.Context(), running{}, dialling(&receiving{}), agentName, requested(request), files, nil, io.Discard)

	// assert
	require.ErrorContains(t, err, "motd: changed since it was planned")
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestAnAgentThatAsksForFilesOfACopyFromAStageGetsNone(t *testing.T) {
	// arrange
	request := protocol.Copy{Key: "step", Stage: "a", From: []string{"base"}, Sources: []string{"/motd"}, Destination: "/etc/motd"}
	var asked []protocol.Copy
	files := func(copying protocol.Copy) protocol.Files {
		asked = append(asked, copying)

		return func(func(protocol.Entry, io.Reader) error) error { return nil }
	}

	// act
	err := builder.Ask(t.Context(), running{}, dialling(&receiving{}), agentName, requested(request), files, nil, io.Discard)

	// assert
	require.ErrorContains(t, err, "protocol.Send is no answer to protocol.Copy")
	assert.Empty(t, asked)
}

func TestAnAgentThatAnswersOutOfTurnFailsTheBuildWithoutWaitingForTheVM(t *testing.T) {
	// arrange
	request := protocol.Copy{Key: "step", Stage: "a", From: []string{"base"}, Sources: []string{"/motd"}, Destination: "/etc/motd"}
	start := time.Now()

	// act
	err := builder.Ask(t.Context(), running{}, dialling(&receiving{}), agentName, requested(request), nil, nil, io.Discard)

	// assert
	require.ErrorIs(t, err, protocol.ErrNoAnswer)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

// running is a VM that keeps running.
type running struct{}

func (running) Done() <-chan struct{} { return nil }

func (running) Err() error { return nil }

func (running) WithoutKVM() bool { return false }

// onTCG is a VM that keeps running on a host without KVM.
type onTCG struct {
	running
}

func (onTCG) WithoutKVM() bool { return true }

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

func (stopped) WithoutKVM() bool { return false }

// dying is a VM that dies while its agent works on a request. Its
// connection breaks first and the VM counts as stopped a moment later, as
// when QEMU is killed.
type dying struct {
	done   chan struct{}
	reason error
}

func (d *dying) Done() <-chan struct{} { return d.done }

func (d *dying) Err() error { return d.reason }

func (*dying) WithoutKVM() bool { return false }

func (d *dying) dial() (io.ReadWriteCloser, error) {
	host, agent := net.Pipe()
	go func() {
		_, _ = protocol.New(agentName, agent, agent).Receive()
		_ = agent.Close()
		time.Sleep(50 * time.Millisecond)
		close(d.done)
	}()

	return host, nil
}

// unreachable dials an agent that never listens.
func unreachable() (io.ReadWriteCloser, error) {
	return nil, errors.New("connection reset by peer")
}

// booting fails the first dials, the way dialling a VM fails before its
// agent listens, and dials after that.
func booting(fails int, dial builder.Dial) builder.Dial {
	return func() (io.ReadWriteCloser, error) {
		if fails > 0 {
			fails--

			return unreachable()
		}

		return dial()
	}
}

// dialling hands out connections to an agent that does its work with the
// runner, one request per connection as the real one answers.
func dialling(runner protocol.Runner) builder.Dial {
	return func() (io.ReadWriteCloser, error) {
		host, agent := net.Pipe()
		go func() {
			_ = protocol.New(agentName, agent, agent).Serve(runner)
			_ = agent.Close()
		}()

		return host, nil
	}
}

// counted dials and counts the dials.
func counted(dial builder.Dial) (builder.Dial, *int) {
	dials := 0
	return func() (io.ReadWriteCloser, error) {
		dials++

		return dial()
	}, &dials
}

// oneImport is a build that only brings in its base image.
func oneImport() []build.Request {
	return requested(protocol.Import{Key: "base", Digest: "sha256:aaaa"})
}

// requested are the requests of a build that asks the agent these messages.
func requested(messages ...protocol.Message) []build.Request {
	requests := make([]build.Request, 0, len(messages))
	for _, message := range messages {
		requests = append(requests, build.Request{Message: message})
	}

	return requests
}

// messagesOf are what the agent is asked for the requests.
func messagesOf(requests []build.Request) []protocol.Message {
	messages := make([]protocol.Message, 0, len(requests))
	for _, request := range requests {
		messages = append(messages, request.Message)
	}

	return messages
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

func (r *recording) Disk(_ context.Context, request protocol.Disk, _ io.Writer) error {
	r.note(request)

	return nil
}

func (r *recording) Fetch(_ context.Context, request protocol.Fetch, _ protocol.Pieces, _ io.Writer) error {
	r.note(request)

	return nil
}

func (r *recording) Copy(_ context.Context, request protocol.Copy, _ protocol.Entries, _ io.Writer) error {
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

func (s saying) Disk(_ context.Context, _ protocol.Disk, out io.Writer) error {
	_, err := io.WriteString(out, s.words)

	return err
}

func (s saying) Fetch(_ context.Context, _ protocol.Fetch, _ protocol.Pieces, out io.Writer) error {
	_, err := io.WriteString(out, s.words)

	return err
}

func (s saying) Copy(_ context.Context, _ protocol.Copy, _ protocol.Entries, out io.Writer) error {
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

func (e exiting) Disk(context.Context, protocol.Disk, io.Writer) error {
	return nil
}

func (e exiting) Fetch(context.Context, protocol.Fetch, protocol.Pieces, io.Writer) error {
	return nil
}

func (e exiting) Copy(context.Context, protocol.Copy, protocol.Entries, io.Writer) error {
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

func (f failing) Disk(context.Context, protocol.Disk, io.Writer) error {
	return f.err
}

func (f failing) Fetch(context.Context, protocol.Fetch, protocol.Pieces, io.Writer) error {
	return f.err
}

func (f failing) Copy(context.Context, protocol.Copy, protocol.Entries, io.Writer) error {
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

func (w waiting) Disk(context.Context, protocol.Disk, io.Writer) error {
	return nil
}

func (w waiting) Fetch(context.Context, protocol.Fetch, protocol.Pieces, io.Writer) error {
	return nil
}

func (w waiting) Copy(context.Context, protocol.Copy, protocol.Entries, io.Writer) error {
	return nil
}

// receiving is an agent that reads the files of a copy and keeps their
// content.
type receiving struct {
	recording

	contents []string
}

func (r *receiving) Copy(_ context.Context, _ protocol.Copy, entries protocol.Entries, _ io.Writer) error {
	for {
		_, content, err := entries.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		read, err := io.ReadAll(content)
		if err != nil {
			return err
		}

		r.contents = append(r.contents, string(read))
	}
}

func (r *recording) Prune(context.Context, protocol.Prune, io.Writer) error { return nil }

func (s saying) Prune(context.Context, protocol.Prune, io.Writer) error { return nil }

func (e exiting) Prune(context.Context, protocol.Prune, io.Writer) error { return nil }

func (f failing) Prune(context.Context, protocol.Prune, io.Writer) error { return nil }

func (w waiting) Prune(context.Context, protocol.Prune, io.Writer) error { return nil }
