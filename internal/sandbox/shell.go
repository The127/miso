package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/runnet"
)

// Run runs the command of a run in a root and answers its exit code.
func Run(ctx context.Context, root string, run protocol.Run, out io.Writer) (int, error) {
	return start(ctx, root, []string{"-c", run.Command}, run.Env, run.Network, streams{out: out})
}

// streams are what the shell of a run reads and writes. It writes to out, or
// to a terminal, which it also reads and has as its controlling terminal.
type streams struct {
	out      io.Writer
	terminal *os.File
}

// start runs /bin/sh with the arguments in a root and answers its exit code.
func start(ctx context.Context, root string, args, env []string, network *protocol.Network, std streams) (int, error) {
	// Go runs nothing between clone and exec, so the agent itself goes first
	// to set up the namespaces, then becomes the shell
	cmd := exec.CommandContext(ctx, "/proc/self/exe", append([]string{root}, args...)...) //nolint:gosec // running what the build file says is what a RUN is
	cmd.Args[0] = helperName
	// the shell is the init of its own PID namespace, so the kernel kills
	// whatever it leaves behind before the wait for it returns. Go makes every
	// mount private in the new mount namespace, so what the run mounts never
	// reaches the agent. A name and network the run changes stay its own
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET, Unshareflags: syscall.CLONE_NEWNS}

	if std.terminal != nil {
		cmd.Stdin, cmd.Stdout, cmd.Stderr = std.terminal, std.terminal, std.terminal
		cmd.SysProcAttr.Setsid = true
		cmd.SysProcAttr.Setctty = true
	} else {
		cmd.Stdout = std.out
		cmd.Stderr = std.out
	}

	// Docker's defaults, and os/exec keeps the last of a key, so the build
	// file's own values win
	cmd.Env = append([]string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=/root",
	}, env...)
	setup, failed, err := os.Pipe()
	if err != nil {
		return 0, err
	}

	defer func() { _ = setup.Close() }()
	// closed again as soon as the helper has it, or the read of setup would
	// never end
	defer func() { _ = failed.Close() }()

	// the helper waits here before it becomes the shell, so the shell never
	// runs before its network is ready
	gate, open, err := os.Pipe()
	if err != nil {
		return 0, err
	}

	defer func() { _ = open.Close() }()
	defer func() { _ = gate.Close() }()

	cmd.ExtraFiles = []*os.File{failed, gate}
	err = cmd.Start()
	_ = failed.Close()
	_ = gate.Close()
	if err != nil {
		return 0, err
	}

	if network != nil {
		removeCard, err := runnet.AddCard(cmd.Process.Pid, network)
		if err != nil {
			// a gate closed unopened stops the helper before its shell
			_ = open.Close()
			_ = cmd.Wait()

			return 0, err
		}

		// after the wait for the shell, which the returns below come after
		defer removeCard()
	}

	// a helper that failed before the gate is not there to read, and says
	// why below
	_, _ = open.Write([]byte{1})
	_ = open.Close()

	// the helper's end closes when the shell starts, so a run that starts
	// reads nothing here
	reason, err := io.ReadAll(setup)
	if err != nil {
		return 0, err
	}

	err = cmd.Wait()
	if len(reason) > 0 {
		return 0, fmt.Errorf("start run: %s", reason)
	}

	return exitCodeOf(ctx, err)
}

// exitCodeOf is the exit code a finished shell answers, from what its wait
// returned.
func exitCodeOf(ctx context.Context, err error) (int, error) {
	// the kill that stops a cancelled command looks like an exit of its own
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if exited, ok := errors.AsType[*exec.ExitError](err); ok {
		// as a shell reports it, 137 for SIGKILL
		if status, ok := exited.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal()), nil
		}

		return exited.ExitCode(), nil
	}

	return 0, err
}
