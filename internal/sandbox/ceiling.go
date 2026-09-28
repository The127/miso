package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/The127/miso/internal/protocol"
)

// Ceiling makes in scratch the layer a run lies under above every other,
// with the resolv.conf of its network in it.
func Ceiling(scratch string, network *protocol.Network) (string, error) {
	ceiling := filepath.Join(scratch, "ceiling")
	etc := filepath.Join(ceiling, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil { //nolint:gosec // an image's /etc is open to all
		return "", err
	}

	var conf strings.Builder
	for _, family := range []protocol.Family{network.IPv4, network.IPv6} {
		if family.Nameserver != "" {
			_, _ = fmt.Fprintf(&conf, "nameserver %s\n", family.Nameserver)
		}
	}

	// apt fetches as a user of its own, who must find the nameserver too
	return ceiling, os.WriteFile(filepath.Join(etc, "resolv.conf"), []byte(conf.String()), 0o644) //nolint:gosec // see above
}
