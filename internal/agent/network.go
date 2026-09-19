package agent

import (
	"encoding/binary"
	"fmt"
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
// the builder's, with the address and gateway of the network.
func addCard(pid int, network *protocol.Network) error {
	address, err := netip.ParsePrefix(network.Address)
	if err != nil {
		return fmt.Errorf("address of the run: %w", err)
	}

	if !address.Addr().Is4() {
		return fmt.Errorf("address of the run: %s is not IPv4", address)
	}

	gateway, err := netip.ParseAddr(network.Gateway)
	if err != nil {
		return fmt.Errorf("gateway of the run: %w", err)
	}

	text, err := os.ReadFile("/sys/class/net/eth0/ifindex")
	if err != nil {
		return err
	}

	parent, err := strconv.ParseUint(strings.TrimSpace(string(text)), 10, 32)
	if err != nil {
		return err
	}

	namespace, err := os.Open(fmt.Sprintf("/proc/%d/ns/net", pid))
	if err != nil {
		return err
	}

	defer func() { _ = namespace.Close() }()

	builder, err := routeSocket(nil)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(builder) }()

	// nothing else in the builder brings its card up, and a card on a card
	// that is down never has a carrier
	if _, err := ask(builder, unix.RTM_SETLINK, 0, link(int32(parent), unix.IFF_UP, unix.IFF_UP)); err != nil { //nolint:gosec // an index fits in an int32
		return fmt.Errorf("bring up the builder's card: %w", err)
	}

	// some kernels look for the name among the builder's cards, where eth0
	// is taken, so the card arrives under a name of its own and is renamed
	arriving := fmt.Sprintf("run%d", pid)
	_, err = ask(builder, unix.RTM_NEWLINK, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link(0, 0, 0),
		attribute(unix.IFLA_IFNAME, []byte(arriving+"\x00")),
		attribute(unix.IFLA_LINK, binary.NativeEndian.AppendUint32(nil, uint32(parent))),
		attribute(unix.IFLA_NET_NS_FD, binary.NativeEndian.AppendUint32(nil, uint32(namespace.Fd()))), //nolint:gosec // a file descriptor fits in 32 bits
		attribute(unix.IFLA_LINKINFO, attribute(unix.IFLA_INFO_KIND, []byte("macvlan"))),
	)
	if err != nil {
		return fmt.Errorf("create card: %w", err)
	}

	run, err := routeSocket(namespace)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(run) }()

	cards, err := ask(run, unix.RTM_GETLINK, unix.NLM_F_DUMP, link(0, 0, 0))
	if err != nil {
		return fmt.Errorf("find card: %w", err)
	}

	for _, card := range cards {
		attributes, err := syscall.ParseNetlinkRouteAttr(&card)
		if err != nil {
			return fmt.Errorf("find card: %w", err)
		}

		for _, a := range attributes {
			if a.Attr.Type != unix.IFLA_IFNAME || strings.TrimRight(string(a.Value), "\x00") != arriving {
				continue
			}

			// the index of the card in the answer's interface header
			index := int32(binary.NativeEndian.Uint32(card.Data[4:])) //nolint:gosec // the kernel writes an int32 there
			if _, err := ask(run, unix.RTM_SETLINK, 0, link(index, 0, 0), attribute(unix.IFLA_IFNAME, []byte("eth0\x00"))); err != nil {
				return fmt.Errorf("name card: %w", err)
			}

			local := address.Addr().As4()
			// the address before the route, which the kernel only takes to a
			// gateway it can reach
			_, err = ask(run, unix.RTM_NEWADDR, unix.NLM_F_CREATE|unix.NLM_F_EXCL, addressHeader(index, address.Bits()),
				attribute(unix.IFA_LOCAL, local[:]),
				attribute(unix.IFA_ADDRESS, local[:]),
			)
			if err != nil {
				return fmt.Errorf("address card %s: %w", address, err)
			}

			if _, err := ask(run, unix.RTM_SETLINK, 0, link(index, unix.IFF_UP, unix.IFF_UP)); err != nil {
				return fmt.Errorf("bring up card: %w", err)
			}

			via := gateway.As4()
			_, err = ask(run, unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_EXCL, routeHeader(),
				attribute(unix.RTA_GATEWAY, via[:]),
				attribute(unix.RTA_OIF, binary.NativeEndian.AppendUint32(nil, uint32(index))), //nolint:gosec // an index is positive
			)
			if err != nil {
				return fmt.Errorf("route via %s: %w", gateway, err)
			}

			return nil
		}
	}

	return fmt.Errorf("find card: %s is not in the run", arriving)
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
	reply := make([]byte, os.Getpagesize())
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
