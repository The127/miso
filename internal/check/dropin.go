package check

import (
	"fmt"

	"github.com/The127/miso/internal/qemu"
)

// dropin pulls the check socket in, because an extra unit is only loaded,
// not started.
func dropin() qemu.Credential {
	return qemu.Credential{Name: "systemd.unit-dropin.sockets.target", Value: fmt.Appendf(nil, `[Unit]
Wants=%s.socket
`, unit)}
}
