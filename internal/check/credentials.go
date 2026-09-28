package check

import (
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/qemu"
)

// unit names the check's socket and service alike, which is how the socket
// finds the service it starts.
const unit = "miso-check"

// Credentials are what the image's systemd is handed for a check.
func Credentials(notifyPort uint32) []qemu.Credential {
	return []qemu.Credential{
		{Name: "vmm.notify_socket", Value: fmt.Appendf(nil, "vsock-stream:%d:%d", unix.VMADDR_CID_HOST, notifyPort)},
		socket(),
		service(),
		dropin(),
	}
}
