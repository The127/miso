package qemu

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// VM is a started builder VM.
type VM struct {
	cid  uint32
	done chan struct{}
	err  error
}

// run starts QEMU and gives the VM that lasts as long as it does.
func run(command *exec.Cmd, cid uint32) (*VM, error) {
	// a QEMU left behind by a miso that was killed would hold its CID and
	// its cache disk for good
	command.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

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
			vm.err = fmt.Errorf("QEMU stopped: %w: %s", err, strings.TrimSpace(refusal.String()))
		}

		close(vm.done)
	}()

	if err := <-started; err != nil {
		return nil, err
	}

	return vm, nil
}

// CID is where the host reaches the VM over vsock.
func (vm *VM) CID() uint32 {
	return vm.cid
}

// Done is closed once the VM has stopped.
func (vm *VM) Done() <-chan struct{} {
	return vm.done
}

// Err is why the VM stopped, once Done is closed.
func (vm *VM) Err() error {
	return vm.err
}
