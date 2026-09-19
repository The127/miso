package agent

import (
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
)

// removeCards removes every card of a run but its loopback, not only its
// own, the run may have made more. The kernel removes the network of a run
// some time after it ended, and until then its cards hold MACs and
// addresses a next run needs.
func removeCards(run *link.Conn) {
	cards, _ := run.Ask(unix.RTM_GETLINK, unix.NLM_F_DUMP, link.CardHeader(0, 0, 0))
	for _, card := range cards {
		if link.Flags(card)&unix.IFF_LOOPBACK != 0 {
			continue
		}

		_, _ = run.Ask(unix.RTM_DELLINK, 0, link.CardHeader(link.Index(card), 0, 0))
	}
}
