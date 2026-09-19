package runnet

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
)

// builderCard finds the builder's card by its MAC, as the host wrote it in
// named, and readies it to carry the cards of runs. It answers its index.
func builderCard(builder *link.Conn, mac []byte, named string) (int32, error) {
	cards, err := builder.Ask(unix.RTM_GETLINK, unix.NLM_F_DUMP, link.CardHeader(0, 0, 0))
	if err != nil {
		return 0, fmt.Errorf("find the builder's card: %w", err)
	}

	for _, card := range cards {
		if !bytes.Equal(link.Value(card, unix.IFLA_ADDRESS), mac) {
			continue
		}

		// it carries the runs and has no address of its own, or IPv6 gives
		// it one the moment it is up. A kernel without IPv6 has nothing to
		// switch off
		disable := "/proc/sys/net/ipv6/conf/" + link.Name(card) + "/disable_ipv6"
		if err := os.WriteFile(disable, []byte("1"), 0o600); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return 0, fmt.Errorf("switch off IPv6 on the builder's card: %w", err)
		}

		// nothing else in the builder brings it up, and a card on a card that
		// is down never has a carrier
		if _, err := builder.Ask(unix.RTM_SETLINK, 0, link.CardHeader(link.Index(card), unix.IFF_UP, unix.IFF_UP)); err != nil {
			return 0, fmt.Errorf("bring up the builder's card: %w", err)
		}

		return link.Index(card), nil
	}

	return 0, fmt.Errorf("the builder has no card %s", named)
}
