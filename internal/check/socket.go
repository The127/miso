package check

import (
	"fmt"

	"github.com/The127/miso/internal/qemu"
)

// shellPort is where the image takes one connection per check.
const shellPort = 5000

// socket starts a shell for each connection. A systemd initrd reads the
// credentials too, but has no vsock yet and would fail the socket.
func socket() qemu.Credential {
	return qemu.Credential{Name: "systemd.extra-unit." + unit + ".socket", Value: fmt.Appendf(nil, `[Unit]
Description=miso check shell
ConditionPathExists=!/etc/initrd-release

[Socket]
ListenStream=vsock::%d
Accept=yes
`, shellPort)}
}
