package builder

import (
	"bufio"
	"fmt"
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
func HostResolving(conf io.Reader) (Resolving, error) {
	var resolving Resolving
	lines := bufio.NewScanner(conf)
	for lines.Scan() {
		address, named := nameserver(lines.Text())
		if !named {
			continue
		}

		resolving.IPv4 = resolving.IPv4 || address.Is4()
		resolving.IPv6 = resolving.IPv6 || address.Is6()
	}

	if err := lines.Err(); err != nil {
		return Resolving{}, fmt.Errorf("read the host's resolv.conf: %w", err)
	}

	return resolving, nil
}

// nameserver is the address a line of resolv.conf names a nameserver at, if
// QEMU would read one there.
func nameserver(line string) (netip.Addr, bool) {
	after, found := strings.CutPrefix(line, "nameserver")
	named := strings.TrimLeft(after, " \t")
	if !found || named == after {
		return netip.Addr{}, false
	}

	// QEMU reads only the word after the keyword, which ends where C's
	// whitespace does
	if end := strings.IndexAny(named, " \t\n\v\f\r"); end >= 0 {
		named = named[:end]
	}

	// QEMU drops the interface after a % from any address, not only an IPv6
	// one
	named, _, _ = strings.Cut(named, "%")

	address, err := netip.ParseAddr(named)

	return address, err == nil
}
