package vsockns

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// ErrNotPrivate is a host on which vsock cannot be kept to miso: its kernel
// has no vsock namespaces, or it refuses a user namespace.
var ErrNotPrivate = errors.New("vsock cannot be kept private on this host")

// childModePath is where the helper makes the namespaces it starts local.
var childModePath = "/proc/sys/net/vsock/child_ns_mode"

// Namespace is a vsock namespace of local mode, held by its helper.
type Namespace struct {
	helper *exec.Cmd
	conn   *os.File
	said   *bufio.Reader
}

// Open sets up a vsock namespace that only miso's VMs share.
func Open() (*Namespace, error) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	conn, theirs := os.NewFile(uintptr(fds[0]), "vsockns"), os.NewFile(uintptr(fds[1]), "vsockns")
	defer func() { _ = theirs.Close() }()

	helper := exec.Command("/proc/self/exe", childModePath)
	helper.Args[0] = helperName
	helper.ExtraFiles = []*os.File{theirs}
	helper.Stderr = os.Stderr
	helper.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:  syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
	}

	if err := helper.Start(); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("%w: %w", ErrNotPrivate, err)
	}

	namespace := &Namespace{helper: helper, conn: conn, said: bufio.NewReader(conn)}

	ready, err := namespace.answer()
	if err == nil && ready != readyWord {
		err = errors.New(ready)
	}

	if err != nil {
		_ = namespace.Close()

		return nil, fmt.Errorf("%w: %w", ErrNotPrivate, err)
	}

	return namespace, nil
}

// Mode is the vsock mode inside the namespace.
func (n *Namespace) Mode() (string, error) {
	if _, err := fmt.Fprintln(n.conn, modeQuestion); err != nil {
		return "", err
	}

	return n.answer()
}

// Close ends the helper, and with it the namespace.
func (n *Namespace) Close() error {
	_ = n.conn.Close()

	return n.helper.Wait()
}

func (n *Namespace) answer() (string, error) {
	line, err := n.said.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(line, "\n"), nil
}
