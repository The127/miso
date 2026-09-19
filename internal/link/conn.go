package link

import (
	"encoding/binary"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// Conn is a route netlink socket in one network namespace.
type Conn struct {
	sock int
}

// Open opens a route netlink socket in a network namespace, or in the
// caller's own one for nil. It stays there whichever thread uses it.
func Open(namespace *os.File) (*Conn, error) {
	var sock int
	err := Within(namespace, func() error {
		var err error
		sock, err = unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.NETLINK_ROUTE)

		return err
	})
	if err != nil {
		return nil, err
	}

	if err := unix.Bind(sock, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		_ = unix.Close(sock)

		return nil, err
	}

	return &Conn{sock: sock}, nil
}

// Close closes the socket.
func (c *Conn) Close() error {
	return unix.Close(c.sock)
}

// Ask sends the kernel a request about the network and answers what came
// back, an error when the kernel refused.
func (c *Conn) Ask(kind, flags uint16, header []byte, attributes ...[]byte) ([]syscall.NetlinkMessage, error) {
	message := append(make([]byte, unix.SizeofNlMsghdr), header...)
	for _, a := range attributes {
		message = append(message, a...)
	}

	binary.NativeEndian.PutUint32(message[0:], uint32(len(message))) //nolint:gosec // a request is far smaller than 4 GiB
	binary.NativeEndian.PutUint16(message[4:], kind)
	binary.NativeEndian.PutUint16(message[6:], unix.NLM_F_REQUEST|unix.NLM_F_ACK|flags)
	binary.NativeEndian.PutUint32(message[8:], 1)

	if err := unix.Sendto(c.sock, message, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return nil, err
	}

	var found []syscall.NetlinkMessage
	// the kernel sends a dump in parts of up to 32 KiB, and cuts off what
	// does not fit
	reply := make([]byte, 32<<10)
	for {
		n, _, err := unix.Recvfrom(c.sock, reply, 0)
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
