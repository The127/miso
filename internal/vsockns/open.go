package vsockns

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// ErrNotPrivate is a host on which vsock cannot be kept to miso: its kernel
// has no vsock namespaces, or it refuses a user namespace.
var ErrNotPrivate = errors.New("vsock cannot be kept private on this host")

// childModePath is where the helper makes the namespaces it starts local.
var childModePath = "/proc/sys/net/vsock/child_ns_mode"

// openPatience is how long Open waits for its helper's first answer.
var openPatience = 10 * time.Second

// devicePath is the host's vhost-vsock device, which the helper opens.
var devicePath = "/dev/vhost-vsock"

// Namespace is a vsock namespace of local mode, held by its helper.
type Namespace struct {
	helper *exec.Cmd
	conn   int

	// a question and its answer share the channel with no other, and none
	// is put once the channel is closed
	asking sync.Mutex
	gone   bool

	closing sync.Once
	closed  error
}

// Open sets up a vsock namespace that only miso's VMs share.
func Open() (*Namespace, error) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	theirs := os.NewFile(uintptr(fds[1]), "vsockns")

	helper := exec.Command("/proc/self/exe", childModePath, devicePath)
	helper.Args[0] = helperName
	// the first of them becomes helperConn
	helper.ExtraFiles = []*os.File{theirs}
	helper.Stderr = os.Stderr
	helper.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:  syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
	}

	err = helper.Start()
	// with miso's copy gone, a helper that ends is the end of the channel
	_ = theirs.Close()

	if err != nil {
		_ = unix.Close(fds[0])

		return nil, fmt.Errorf("%w: %w", ErrNotPrivate, err)
	}

	namespace := &Namespace{helper: helper, conn: fds[0]}

	if err := namespace.ready(); err != nil {
		_ = namespace.Close()

		return nil, err
	}

	return namespace, nil
}

// ready waits for the helper's first answer, for as long as the patience
// lasts.
func (n *Namespace) ready() error {
	answered := make(chan error, 1)
	go func() {
		_, _, err := n.answer()
		answered <- err
	}()

	select {
	case err := <-answered:
		if err != nil {
			return fmt.Errorf("%w: %w", ErrNotPrivate, err)
		}

		return nil
	case <-time.After(openPatience):
		// a helper stuck before its first answer never reads the end of the
		// channel, so the channel is ended under it and the helper with it
		_ = unix.Shutdown(n.conn, unix.SHUT_RDWR)
		_ = n.helper.Process.Kill()
		<-answered

		return fmt.Errorf("the helper of the vsock namespace did not answer within %s", openPatience)
	}
}

// Mode is the vsock mode inside the namespace.
func (n *Namespace) Mode() (string, error) {
	mode, _, err := n.ask(modeQuestion)

	return mode, err
}
