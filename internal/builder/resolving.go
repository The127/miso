package builder

import (
	"bufio"
	"io"
	"net/netip"
	"strings"
)

// Resolving is which address families the host resolves names in. QEMU
// forwards a query only to a resolver of the query's own family.
type Resolving struct {
	IPv4 bool
	IPv6 bool
}

// HostResolving is which families the host resolves names in, read from
// its resolv.conf the way QEMU reads it.
func HostResolving(conf io.Reader) Resolving {
	var resolving Resolving
	lines := bufio.NewScanner(conf)
	for lines.Scan() {
		after, found := strings.CutPrefix(lines.Text(), "nameserver")
		named := strings.TrimLeft(after, " \t")
		if !found || named == after {
			continue
		}

		// QEMU drops the interface after a % from any address, not only an
		// IPv6 one
		named, _, _ = strings.Cut(named, "%")

		address, err := netip.ParseAddr(named)
		if err != nil {
			continue
		}

		resolving.IPv4 = resolving.IPv4 || address.Is4()
		resolving.IPv6 = resolving.IPv6 || address.Is6()
	}

	return resolving
}
