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

// AddressHeader is the header of a request about an address of a card, in
// a family, IFA_F_NODAD and the like among the flags.
func AddressHeader(family byte, index int32, bits int, flags byte) []byte {
	header := []byte{family, byte(bits), flags, unix.RT_SCOPE_UNIVERSE} //nolint:gosec // a prefix length is at most 128

	return binary.NativeEndian.AppendUint32(header, uint32(index)) //nolint:gosec // an index is positive
}

// TrafficHeader is the header of a request about how a card treats its
// traffic: the card's index, the handle, the parent and the info of a
// queue or filter.
func TrafficHeader(index int32, handle, parent, info uint32) []byte {
	header := binary.NativeEndian.AppendUint32([]byte{unix.AF_UNSPEC, 0, 0, 0}, uint32(index)) //nolint:gosec // the header holds an int32
	header = binary.NativeEndian.AppendUint32(header, handle)
	header = binary.NativeEndian.AppendUint32(header, parent)

	return binary.NativeEndian.AppendUint32(header, info)
}

// RouteHeader is the header of a request for the default route of a
// family.
func RouteHeader(family byte) []byte {
	header := []byte{family, 0, 0, 0, unix.RT_TABLE_MAIN, unix.RTPROT_BOOT, unix.RT_SCOPE_UNIVERSE, unix.RTN_UNICAST}

	return binary.NativeEndian.AppendUint32(header, 0)
}
