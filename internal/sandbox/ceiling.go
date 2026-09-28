package sandbox

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/The127/miso/internal/protocol"
)

// Ceiling makes in scratch the layer a run lies under above every other,
// with the resolv.conf of its network in it. Its etc looks like the one of
// the layers below, top first, because overlay copies it into the run's
// layer when the run writes there.
func Ceiling(scratch string, below []string, network *protocol.Network) (string, error) {
	ceiling := filepath.Join(scratch, "ceiling")
	etc := filepath.Join(ceiling, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil { //nolint:gosec // an image's /etc is open to all
		return "", err
	}

	for _, layer := range below {
		info, err := os.Lstat(filepath.Join(layer, "etc"))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return "", err
		}

		if err := os.Chmod(etc, info.Mode().Perm()); err != nil {
			return "", err
		}

		owner, _ := info.Sys().(*syscall.Stat_t)
		if err := os.Lchown(etc, int(owner.Uid), int(owner.Gid)); err != nil {
			return "", err
		}

		break
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
