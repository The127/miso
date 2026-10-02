package check

import "github.com/The127/miso/internal/qemu"

// the parts of a check boot the tests drive one by one
var (
	Credentials = credentials
	Booted      = booted
	Run         = run
)

type (
	Notices = noticeListener
	Conn    = shellConn
)

// BootMethod is how the boot starts its machine.
func (b Boot) BootMethod() (qemu.Boot, error) { return b.method() }

// Machine is the machine the boot starts.
func (b Boot) Machine(method qemu.Boot, notifyPort uint32) qemu.Machine {
	return b.machine(method, notifyPort)
}
