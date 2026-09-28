package check

import "github.com/The127/miso/internal/qemu"

// dropin pulls the check socket in, because an extra unit is only loaded,
// not started.
func dropin() qemu.Credential {
	return qemu.Credential{Name: "systemd.unit-dropin.sockets.target", Value: []byte(`[Unit]
Wants=miso-check.socket
`)}
}
