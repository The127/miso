package runnet

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
	"github.com/The127/miso/internal/protocol"
)

// AddCard gives the run of a process a network card of its own on top of
// the builder's, with the address and gateway of the network, and answers
// how to remove it once the run has ended.
func AddCard(pid int, network *protocol.Network) (func(), error) {
	wanted, err := readSettings(network)
	if err != nil {
		return nil, err
	}

	namespace, err := os.Open(fmt.Sprintf("/proc/%d/ns/net", pid))
	if err != nil {
		return nil, err
	}

	defer func() { _ = namespace.Close() }()

	builder, err := link.Open(nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = builder.Close() }()

	parent, err := builderCard(builder, wanted.card, network.Card)
	if err != nil {
		return nil, err
	}

	if err := fenceHost(builder, parent, wanted.ipv4.gateway); err != nil {
		return nil, err
	}

	// some kernels look for the name among the builder's cards, where eth0
	// is taken, so the card arrives under a name of its own and is renamed
	arriving := fmt.Sprintf("run%d", pid)
	if err := createCard(builder, parent, namespace, arriving, wanted.mac()); err != nil {
		return nil, err
	}

	// open until the cards are gone, it holds the run's network
	run, err := link.Open(namespace)
	if err != nil {
		return nil, err
	}

	index, err := configureCard(run, namespace, arriving, wanted)
	if err != nil {
		// a card that failed half way goes at once, or it holds the MAC and
		// address until the kernel removes the run's network
		if index != 0 {
			_, _ = run.Ask(unix.RTM_DELLINK, 0, link.CardHeader(index, 0, 0))
		}

		_ = run.Close()

		return nil, err
	}

	return func() {
		removeCards(run)
		_ = run.Close()
	}, nil
}
