package main

import (
	"errors"
	"io/fs"
	"os"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// hostNetwork is the builder's card and the network its runs get, with a
// nameserver in the families the host resolves in.
func hostNetwork() (qemu.Card, protocol.Network, error) {
	var resolving builder.Resolving

	conf, err := os.Open("/etc/resolv.conf")
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// QEMU finds no resolver either
	case err != nil:
		return qemu.Card{}, protocol.Network{}, err
	default:
		defer func() { _ = conf.Close() }()

		resolving, err = builder.HostResolving(conf)
		if err != nil {
			return qemu.Card{}, protocol.Network{}, err
		}
	}

	card, network := builder.Network(resolving)

	return card, network, nil
}
