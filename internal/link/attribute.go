package link

import (
	"encoding/binary"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// Attribute is a netlink attribute: its length, its type, the value, padded
// to four bytes.
func Attribute(kind uint16, value []byte) []byte {
	length := unix.SizeofRtAttr + len(value)
	found := binary.NativeEndian.AppendUint16(nil, uint16(length)) //nolint:gosec // an attribute is far smaller than 64 KiB
	found = binary.NativeEndian.AppendUint16(found, kind)
	found = append(found, value...)

	return append(found, make([]byte, (4-length%4)%4)...)
}

// Value is the value of an attribute of a card as the kernel described it,
// nil when it has none. It reads no more than lengths and types, which
// every kernel writes alike.
func Value(card syscall.NetlinkMessage, kind uint16) []byte {
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

// Index is the index of a card as the kernel described it.
func Index(card syscall.NetlinkMessage) int32 {
	return int32(binary.NativeEndian.Uint32(card.Data[4:])) //nolint:gosec // the kernel writes an int32 there
}

// Name is the name of a card as the kernel described it.
func Name(card syscall.NetlinkMessage) string {
	return strings.TrimRight(string(Value(card, unix.IFLA_IFNAME)), "\x00")
}

// Flags are the flags of a card as the kernel described it, IFF_UP and the
// like.
func Flags(card syscall.NetlinkMessage) uint32 {
	return binary.NativeEndian.Uint32(card.Data[8:])
}
