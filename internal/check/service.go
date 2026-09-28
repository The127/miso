package check

import (
	"fmt"

	"github.com/The127/miso/internal/qemu"
)

// exitMarker starts the last line of a check's output, the exit code
// follows it.
const exitMarker = "miso-exit"

// service runs the check the host writes to the connection. systemd would
// expand a single $ itself, and $? must be read before the echo resets it.
// The echo puts the marker on a line of its own after output that does not
// end in one.
func service() qemu.Credential {
	return qemu.Credential{Name: "systemd.extra-unit.miso-check@.service", Value: fmt.Appendf(nil, `[Unit]
Description=miso check

[Service]
ExecStart=/bin/sh -c '/bin/sh; code=$$?; echo; echo "%s $$code"'
StandardInput=socket
StandardOutput=socket
StandardError=socket
`, exitMarker)}
}
