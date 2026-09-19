package link

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

// CardHeader is the header of a request about a card: its index, and the
// flags to set among those to change.
func CardHeader(index int32, flags, change uint32) []byte {
	header := make([]byte, unix.SizeofIfInfomsg)
	binary.NativeEndian.PutUint32(header[4:], uint32(index)) //nolint:gosec // the header holds an int32
	binary.NativeEndian.PutUint32(header[8:], flags)
	binary.NativeEndian.PutUint32(header[12:], change)

	return header
}

// AddressHeader is the header of a request about an IPv4 address of a
// card.
func AddressHeader(index int32, bits int) []byte {
	header := []byte{unix.AF_INET, byte(bits), 0, unix.RT_SCOPE_UNIVERSE} //nolint:gosec // a prefix length is at most 32

	return binary.NativeEndian.AppendUint32(header, uint32(index)) //nolint:gosec // an index is positive
}

// RouteHeader is the header of a request for the default IPv4 route.
func RouteHeader() []byte {
	header := []byte{unix.AF_INET, 0, 0, 0, unix.RT_TABLE_MAIN, unix.RTPROT_BOOT, unix.RT_SCOPE_UNIVERSE, unix.RTN_UNICAST}

	return binary.NativeEndian.AppendUint32(header, 0)
}
