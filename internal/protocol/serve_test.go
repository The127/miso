package protocol_test

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

	"github.com/The127/miso/internal/protocol"
)

type runner struct {
	writes string
	code   int
	err    error
}

func (r runner) Run(_ context.Context, _ protocol.Run, out io.Writer) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.writes == "" {
		return r.code, nil
	}

	_, err := io.WriteString(out, r.writes)

	return r.code, err
}

func (r runner) Import(_ context.Context, _ protocol.Import, out io.Writer) error {
	if r.err != nil || r.writes == "" {
		return r.err
	}

	_, err := io.WriteString(out, r.writes)

	return err
}

func (r runner) Copy(_ context.Context, _ protocol.Copy, _ protocol.Entries, _ io.Writer) error {
	return nil
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

func TestAnImportThatWorksIsDone(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Import{Key: "abc", Digest: "sha256:def"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{})

	// assert
	require.NoError(t, err)
	done, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Done{}, done)
}

func TestTheOutputOfAnImportReachesTheHost(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Import{Key: "abc", Digest: "sha256:def"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{writes: "copied 1 GB\n"})

	// assert
	require.NoError(t, err)
	output, err := host.Receive()
	require.NoError(t, err)
	done, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Output{Bytes: []byte("copied 1 GB\n")}, output)
	assert.Equal(t, protocol.Done{}, done)
}

func TestAnImportErrorIsFailed(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Import{Key: "abc", Digest: "sha256:def"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{err: errors.New("no root partition")})

	// assert
	require.NoError(t, err)
	failed, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Failed{Reason: "no root partition"}, failed)
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

func (w *watched) Run(context.Context, protocol.Run, io.Writer) (int, error) {
	w.ran = true

	return 0, nil
}

func (w *watched) Import(context.Context, protocol.Import, io.Writer) error {
	return nil
}

func (w *watched) Copy(context.Context, protocol.Copy, protocol.Entries, io.Writer) error {
	return nil
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

type waiting struct {
	cancelled bool
}

func (w *waiting) Run(ctx context.Context, _ protocol.Run, _ io.Writer) (int, error) {
	select {
	case <-ctx.Done():
		w.cancelled = true

		return 0, ctx.Err()
	case <-time.After(5 * time.Second):
		return 0, errors.New("never cancelled")
	}
}

func (w *waiting) Import(ctx context.Context, _ protocol.Import, _ io.Writer) error {
	_, err := w.Run(ctx, protocol.Run{}, io.Discard)

	return err
}

func (w *waiting) Copy(ctx context.Context, _ protocol.Copy, _ protocol.Entries, _ io.Writer) error {
	return w.Import(ctx, protocol.Import{}, io.Discard)
}

func TestClosingTheConnectionCancelsTheRun(t *testing.T) {
	// arrange
	requests, hostEnd := io.Pipe()
	var replies bytes.Buffer
	host := protocol.New("miso 1.2.0", bytes.NewReader(nil), hostEnd)
	agent := protocol.New("miso 1.2.0", requests, &replies)
	runner := &waiting{}
	go func() {
		_ = host.Send(protocol.Run{Command: "sleep infinity"})
		_ = hostEnd.Close()
	}()

	// act
	_ = agent.Serve(runner)

	// assert
	assert.True(t, runner.cancelled)
}

func TestClosingTheConnectionCancelsTheImport(t *testing.T) {
	// arrange
	requests, hostEnd := io.Pipe()
	var replies bytes.Buffer
	host := protocol.New("miso 1.2.0", bytes.NewReader(nil), hostEnd)
	agent := protocol.New("miso 1.2.0", requests, &replies)
	runner := &waiting{}
	go func() {
		_ = host.Send(protocol.Import{Key: "abc", Digest: "sha256:def"})
		_ = hostEnd.Close()
	}()

	// act
	_ = agent.Serve(runner)

	// assert
	assert.True(t, runner.cancelled)
}

func TestAWriteLongerThanAMessageReachesTheHostWhole(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Run{Command: "cat big"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	big := strings.Repeat("x", 2*protocol.MaxMessage)

	// act
	err := agent.Serve(runner{writes: big})

	// assert
	require.NoError(t, err)
	var received strings.Builder
	for {
		message, err := host.Receive()
		require.NoError(t, err)
		output, isOutput := message.(protocol.Output)
		if !isOutput {
			assert.Equal(t, protocol.Done{}, message)

			break
		}

		received.Write(output.Bytes)
	}

	assert.Equal(t, big, received.String())
}

func TestACopyWhoseLayerIsCachedIsDoneWithoutAskingForEntries(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Copy{Key: "abc", Destination: "/etc/motd"}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(runner{})

	// assert
	require.NoError(t, err)
	done, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Done{}, done)
}

// reading is an agent that reads every entry of a copy and keeps them with
// their content.
type reading struct {
	entries  []protocol.Entry
	contents []string
}

func (r *reading) Run(context.Context, protocol.Run, io.Writer) (int, error) {
	return 0, nil
}

func (r *reading) Import(context.Context, protocol.Import, io.Writer) error {
	return nil
}

func (r *reading) Copy(_ context.Context, _ protocol.Copy, entries protocol.Entries, _ io.Writer) error {
	for {
		entry, content, err := entries.Next()
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

		r.entries = append(r.entries, entry)
		r.contents = append(r.contents, string(read))
	}
}

func TestACopyAsksForItsEntriesAndReadsThemUntilSent(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	entry := protocol.Entry{Kind: "file", Path: ".", Mode: 0o644, Size: 6}
	require.NoError(t, host.Send(protocol.Copy{Key: "abc", Sources: []string{"motd"}, Destination: "/etc/motd"}))
	require.NoError(t, host.SendEntry(entry, strings.NewReader("hello\n")))
	require.NoError(t, host.Send(protocol.Sent{}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)
	runner := &reading{}

	// act
	err := agent.Serve(runner)

	// assert
	require.NoError(t, err)
	send, err := host.Receive()
	require.NoError(t, err)
	done, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Send{}, send)
	assert.Equal(t, protocol.Done{}, done)
	assert.Equal(t, []protocol.Entry{entry}, runner.entries)
	assert.Equal(t, []string{"hello\n"}, runner.contents)
}

// breaking is an agent that fails after the first entry of a copy.
type breaking struct {
	err error
}

func (b breaking) Run(context.Context, protocol.Run, io.Writer) (int, error) {
	return 0, nil
}

func (b breaking) Import(context.Context, protocol.Import, io.Writer) error {
	return nil
}

func (b breaking) Copy(_ context.Context, _ protocol.Copy, entries protocol.Entries, _ io.Writer) error {
	if _, _, err := entries.Next(); err != nil {
		return err
	}

	return b.err
}

func TestACopyThatFailsReadsTheRestOfItsEntriesBeforeItSaysSo(t *testing.T) {
	// arrange
	var requests, replies bytes.Buffer
	host := protocol.New("miso 1.2.0", &replies, &requests)
	require.NoError(t, host.Send(protocol.Copy{Key: "abc", Sources: []string{"etc"}, Destination: "/etc"}))
	require.NoError(t, host.SendEntry(protocol.Entry{Kind: "file", Path: "motd", Mode: 0o644, Size: 6}, strings.NewReader("hello\n")))
	require.NoError(t, host.SendEntry(protocol.Entry{Kind: "file", Path: "issue", Mode: 0o644, Size: 4}, strings.NewReader("hey\n")))
	require.NoError(t, host.Send(protocol.Sent{}))
	agent := protocol.New("miso 1.2.0", &requests, &replies)

	// act
	err := agent.Serve(breaking{err: errors.New("disk full")})

	// assert
	require.NoError(t, err)
	send, err := host.Receive()
	require.NoError(t, err)
	failed, err := host.Receive()
	require.NoError(t, err)
	assert.Equal(t, protocol.Send{}, send)
	assert.Equal(t, protocol.Failed{Reason: "disk full"}, failed)
	assert.Zero(t, requests.Len())
}
