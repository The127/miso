package builder

import (
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// mac is the builder's card's, 52:54:00 as QEMU's own and then mis.
const mac = "52:54:00:6d:69:73"

// Network is the builder VM's card on QEMU's user network and the network a
// run gets through it, from one set of numbers so the two agree.
func Network() (qemu.Card, protocol.Network) {
	return qemu.Card{MAC: mac}, protocol.Network{Card: mac}
}
