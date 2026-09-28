package qemu

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// VM is a started builder VM.
type VM struct {
	cid        uint32
	done       chan struct{}
	err        error
	withoutKVM bool
	port       *os.File
}

// run starts QEMU and gives the VM that lasts as long as it does.
func run(command *exec.Cmd, cid uint32) (*VM, error) {
	// a QEMU left behind by a miso that was killed would hold its CID and
	// its cache disk for good. In a process group of its own it never sees
	// a Ctrl-C meant for miso, which stops it by its context instead
	command.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL, Setpgid: true}

	// QEMU says on its standard error why it cannot run a machine
	var refusal bytes.Buffer
	command.Stderr = &refusal

	vm := &VM{cid: cid, done: make(chan struct{})}
	started := make(chan error)

	go func() {
		// the kernel sends the death signal when the thread that started QEMU
		// ends, not the process, so that thread is held until QEMU is gone
		runtime.LockOSThread()

		if err := command.Start(); err != nil {
			started <- err

			return
		}

		started <- nil

		if err := command.Wait(); err != nil {
			vm.err = stopped(err, refusal.String())
		}

		close(vm.done)
	}()

	if err := <-started; err != nil {
		return nil, err
	}

	return vm, nil
}

func stopped(err error, said string) error {
	said = strings.TrimSpace(said)
	if said == "" {
		return fmt.Errorf("QEMU stopped: %w", err)
	}

	return fmt.Errorf("QEMU stopped: %w: %s", err, said)
}

// CID is where the host reaches the VM over vsock.
func (vm *VM) CID() uint32 {
	return vm.cid
}

// Port is the host's end of the machine's virtio port for its agent. It is
// nil on a host with vsock.
func (vm *VM) Port() *os.File {
	return vm.port
}

// WithoutKVM is whether the VM runs on TCG, because the host has no KVM.
func (vm *VM) WithoutKVM() bool {
	return vm.withoutKVM
}

// Done is closed once the VM has stopped.
func (vm *VM) Done() <-chan struct{} {
	return vm.done
}

// Err is why the VM stopped, once Done is closed.
func (vm *VM) Err() error {
	return vm.err
}
