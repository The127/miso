package agent

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
)

// addCard gives the run of a process a network card of its own on top of
// the builder's, with the address and gateway of the network, and answers
// how to remove it once the run has ended.
func addCard(pid int, network *protocol.Network) (func(), error) {
	address, err := netip.ParsePrefix(network.Address)
	if err != nil {
		return nil, fmt.Errorf("address of the run: %w", err)
	}

	if !address.Addr().Is4() {
		return nil, fmt.Errorf("address of the run: %s is not IPv4", address)
	}

	gateway, err := netip.ParseAddr(network.Gateway)
	if err != nil {
		return nil, fmt.Errorf("gateway of the run: %w", err)
	}

	if !gateway.Is4() {
		return nil, fmt.Errorf("gateway of the run: %s is not IPv4", gateway)
	}

	// the card is known by the MAC the host gave it, its name and place
	// differ between builders
	var mac []byte
	for _, part := range strings.Split(network.Card, ":") {
		octet, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("card of the run %s: %w", network.Card, err)
		}

		mac = append(mac, byte(octet))
	}

	namespace, err := os.Open(fmt.Sprintf("/proc/%d/ns/net", pid))
	if err != nil {
		return nil, err
	}

	defer func() { _ = namespace.Close() }()

	builder, err := routeSocket(nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = unix.Close(builder) }()

	builderCards, err := ask(builder, unix.RTM_GETLINK, unix.NLM_F_DUMP, link(0, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("find the builder's card: %w", err)
	}

	var parent int32
	var name string
	for _, card := range builderCards {
		if bytes.Equal(cardAttribute(card, unix.IFLA_ADDRESS), mac) {
			parent = int32(binary.NativeEndian.Uint32(card.Data[4:])) //nolint:gosec // the kernel writes an int32 there
			name = strings.TrimRight(string(cardAttribute(card, unix.IFLA_IFNAME)), "\x00")
		}
	}

	if parent == 0 {
		return nil, fmt.Errorf("the builder has no card %s", network.Card)
	}

	// the builder's card carries the runs and has no address of its own, or
	// IPv6 gives it one the moment it is up. A kernel without IPv6 has
	// nothing to switch off
	disable := "/proc/sys/net/ipv6/conf/" + name + "/disable_ipv6"
	if err := os.WriteFile(disable, []byte("1"), 0o600); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("switch off IPv6 on the builder's card: %w", err)
	}

	// nothing else in the builder brings its card up, and a card on a card
	// that is down never has a carrier
	if _, err := ask(builder, unix.RTM_SETLINK, 0, link(parent, unix.IFF_UP, unix.IFF_UP)); err != nil {
		return nil, fmt.Errorf("bring up the builder's card: %w", err)
	}

	// some kernels look for the name among the builder's cards, where eth0
	// is taken, so the card arrives under a name of its own and is renamed
	arriving := fmt.Sprintf("run%d", pid)
	// the MAC follows from the address, so a run sees the same one every
	// time, and two runs on one address collide loudly instead of taking
	// turns in the gateway's table. 02 is a MAC of our own making
	local := address.Addr().As4()
	own := append([]byte{0x02, 0x00}, local[:]...)
	_, err = ask(builder, unix.RTM_NEWLINK, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link(0, 0, 0),
		attribute(unix.IFLA_IFNAME, []byte(arriving+"\x00")),
		attribute(unix.IFLA_ADDRESS, own),
		attribute(unix.IFLA_LINK, binary.NativeEndian.AppendUint32(nil, uint32(parent))),
		attribute(unix.IFLA_NET_NS_FD, binary.NativeEndian.AppendUint32(nil, uint32(namespace.Fd()))), //nolint:gosec // a file descriptor fits in 32 bits
		attribute(unix.IFLA_LINKINFO, attribute(unix.IFLA_INFO_KIND, []byte("macvlan"))),
	)
	if err != nil {
		return nil, fmt.Errorf("create card: %w", err)
	}

	run, err := routeSocket(namespace)
	if err != nil {
		return nil, err
	}

	// kept open on success, it holds the run's network until the card is
	// gone. A card that failed half way is gone at once, or it holds the
	// MAC and address until the kernel removes the run's network
	kept := false
	var index int32
	defer func() {
		if kept {
			return
		}

		if index != 0 {
			_, _ = ask(run, unix.RTM_DELLINK, 0, link(index, 0, 0))
		}

		_ = unix.Close(run)
	}()

	cards, err := ask(run, unix.RTM_GETLINK, unix.NLM_F_DUMP, link(0, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("find card: %w", err)
	}

	for _, card := range cards {
		if strings.TrimRight(string(cardAttribute(card, unix.IFLA_IFNAME)), "\x00") != arriving {
			continue
		}

		// the index of the card in the answer's interface header
		index = int32(binary.NativeEndian.Uint32(card.Data[4:])) //nolint:gosec // the kernel writes an int32 there
		if _, err := ask(run, unix.RTM_SETLINK, 0, link(index, 0, 0), attribute(unix.IFLA_IFNAME, []byte("eth0\x00"))); err != nil {
			return nil, fmt.Errorf("name card: %w", err)
		}

		// the address before the route, which the kernel only takes to a
		// gateway it can reach
		_, err = ask(run, unix.RTM_NEWADDR, unix.NLM_F_CREATE|unix.NLM_F_EXCL, addressHeader(index, address.Bits()),
			attribute(unix.IFA_LOCAL, local[:]),
			attribute(unix.IFA_ADDRESS, local[:]),
		)
		if err != nil {
			return nil, fmt.Errorf("address card %s: %w", address, err)
		}

		if _, err := ask(run, unix.RTM_SETLINK, 0, link(index, unix.IFF_UP, unix.IFF_UP)); err != nil {
			return nil, fmt.Errorf("bring up card: %w", err)
		}

		via := gateway.As4()
		_, err = ask(run, unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_EXCL, routeHeader(),
			attribute(unix.RTA_GATEWAY, via[:]),
			attribute(unix.RTA_OIF, binary.NativeEndian.AppendUint32(nil, uint32(index))), //nolint:gosec // an index is positive
		)
		if err != nil {
			return nil, fmt.Errorf("route via %s: %w", gateway, err)
		}

		kept = true
		// the kernel removes the network of a run some time after it ended,
		// and until then its card holds the MAC and address the next run
		// needs
		// every card, not only its own, the run may have made more
		return func() {
			cards, _ := ask(run, unix.RTM_GETLINK, unix.NLM_F_DUMP, link(0, 0, 0))
			for _, card := range cards {
				if binary.NativeEndian.Uint32(card.Data[8:])&unix.IFF_LOOPBACK != 0 {
					continue
				}

				_, _ = ask(run, unix.RTM_DELLINK, 0, link(int32(binary.NativeEndian.Uint32(card.Data[4:])), 0, 0)) //nolint:gosec // the kernel writes an int32 there
			}

			_ = unix.Close(run)
		}, nil
	}

	return nil, fmt.Errorf("find card: %s is not in the run", arriving)
}

// routeSocket opens a route netlink socket in a network namespace, or in
// the agent's own one for nil. It stays there whichever thread uses it.
func routeSocket(namespace *os.File) (int, error) {
	type opened struct {
		sock int
		err  error
	}

	done := make(chan opened)
	go func() {
		if namespace != nil {
			// never unlocked, so the thread that moved ends with this goroutine
			runtime.LockOSThread()
			if err := unix.Setns(int(namespace.Fd()), unix.CLONE_NEWNET); err != nil {
				done <- opened{-1, fmt.Errorf("enter the run's network: %w", err)}

				return
			}
		}

		sock, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.NETLINK_ROUTE)
		done <- opened{sock, err}
	}()

	found := <-done
	if found.err != nil {
		return -1, found.err
	}

	if err := unix.Bind(found.sock, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		_ = unix.Close(found.sock)

		return -1, err
	}

	return found.sock, nil
}

// ask sends the kernel a request about the network and answers what came
// back, an error when the kernel refused.
func ask(sock int, kind, flags uint16, header []byte, attributes ...[]byte) ([]syscall.NetlinkMessage, error) {
	message := append(make([]byte, unix.SizeofNlMsghdr), header...)
	for _, a := range attributes {
		message = append(message, a...)
	}

	binary.NativeEndian.PutUint32(message[0:], uint32(len(message))) //nolint:gosec // a request is far smaller than 4 GiB
	binary.NativeEndian.PutUint16(message[4:], kind)
	binary.NativeEndian.PutUint16(message[6:], unix.NLM_F_REQUEST|unix.NLM_F_ACK|flags)
	binary.NativeEndian.PutUint32(message[8:], 1)

	if err := unix.Sendto(sock, message, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return nil, err
	}

	var found []syscall.NetlinkMessage
	// the kernel sends a dump in parts of up to 32 KiB, and cuts off what
	// does not fit
	reply := make([]byte, 32<<10)
	for {
		n, _, err := unix.Recvfrom(sock, reply, 0)
		if err != nil {
			return nil, err
		}

		answers, err := syscall.ParseNetlinkMessage(reply[:n])
		if err != nil {
			return nil, err
		}

		for _, answer := range answers {
			if answer.Header.Type == unix.NLMSG_DONE {
				return found, nil
			}

			if answer.Header.Type != unix.NLMSG_ERROR {
				found = append(found, answer)

				continue
			}

			// the kernel acknowledges with a negated errno, zero for success
			if code := int32(binary.NativeEndian.Uint32(answer.Data)); code != 0 { //nolint:gosec // the kernel writes an int32 there
				return nil, syscall.Errno(-code) //nolint:gosec // an errno is small and positive
			}

			return found, nil
		}
	}
}

// cardAttribute is the value of an attribute of a card as the kernel
// described it, nil when it has none. It reads no more than lengths and
// types, which every kernel writes alike.
func cardAttribute(card syscall.NetlinkMessage, kind uint16) []byte {
	rest := card.Data[min(unix.SizeofIfInfomsg, len(card.Data)):]
	for len(rest) >= unix.SizeofRtAttr {
		length := int(binary.NativeEndian.Uint16(rest))
		if length < unix.SizeofRtAttr || length > len(rest) {
			return nil
		}

		if binary.NativeEndian.Uint16(rest[2:])&^(unix.NLA_F_NESTED|unix.NLA_F_NET_BYTEORDER) == kind {
			return rest[unix.SizeofRtAttr:length]
		}

		rest = rest[min((length+3)&^3, len(rest)):]
	}

	return nil
}

// link is the header of a request about a card: its index, and the flags
// to set among those to change.
func link(index int32, flags, change uint32) []byte {
	header := make([]byte, unix.SizeofIfInfomsg)
	binary.NativeEndian.PutUint32(header[4:], uint32(index)) //nolint:gosec // the header holds an int32
	binary.NativeEndian.PutUint32(header[8:], flags)
	binary.NativeEndian.PutUint32(header[12:], change)

	return header
}

// addressHeader is the header of a request about an IPv4 address of a card.
func addressHeader(index int32, bits int) []byte {
	header := []byte{unix.AF_INET, byte(bits), 0, unix.RT_SCOPE_UNIVERSE} //nolint:gosec // a prefix length is at most 32

	return binary.NativeEndian.AppendUint32(header, uint32(index)) //nolint:gosec // an index is positive
}

// routeHeader is the header of a request for the default IPv4 route.
func routeHeader() []byte {
	header := []byte{unix.AF_INET, 0, 0, 0, unix.RT_TABLE_MAIN, unix.RTPROT_BOOT, unix.RT_SCOPE_UNIVERSE, unix.RTN_UNICAST}

	return binary.NativeEndian.AppendUint32(header, 0)
}

// attribute is a netlink attribute: its length, its type, the value, padded
// to four bytes.
func attribute(kind uint16, value []byte) []byte {
	length := unix.SizeofRtAttr + len(value)
	found := binary.NativeEndian.AppendUint16(nil, uint16(length)) //nolint:gosec // an attribute is far smaller than 64 KiB
	found = binary.NativeEndian.AppendUint16(found, kind)
	found = append(found, value...)

	return append(found, make([]byte, (4-length%4)%4)...)
}
