package tree

import (
	"bytes"
	"os"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
)

// xattrs gives a copy the extended attributes of what it was copied from,
// never following a link.
func (c copier) xattrs(name string) error {
	var found []place.Xattr

	err := at(c.from, "getxattr", name, func(source string) error {
		var err error

		found, err = read(source)

		return err
	})
	if err != nil {
		return err
	}

	return at(c.to, "setxattr", name, func(target string) error {
		for _, a := range found {
			if err := unix.Lsetxattr(target, a.Name, a.Value, 0); err != nil {
				return err
			}
		}

		return nil
	})
}

// read is every extended attribute of a path with its value, never
// following a link.
func read(path string) ([]place.Xattr, error) {
	return attributes(
		func(names []byte) (int, error) { return unix.Llistxattr(path, names) },
		func(name string, value []byte) (int, error) { return unix.Lgetxattr(path, name, value) },
	)
}

// fileXattrs is every extended attribute of an open file with its value.
func fileXattrs(file *os.File) ([]place.Xattr, error) {
	fd := int(file.Fd())

	return attributes(
		func(names []byte) (int, error) { return unix.Flistxattr(fd, names) },
		func(name string, value []byte) (int, error) { return unix.Fgetxattr(fd, name, value) },
	)
}

// attributes is every extended attribute the list names, with the value
// get reads. Both answer a nil buffer with the size they need.
func attributes(list func(names []byte) (int, error), get func(name string, value []byte) (int, error)) ([]place.Xattr, error) {
	names, err := sized(list)
	if err != nil || len(names) == 0 {
		return nil, err
	}

	var xattrs []place.Xattr

	for name := range bytes.SplitSeq(names, []byte{0}) {
		if len(name) == 0 {
			continue
		}

		value, err := sized(func(value []byte) (int, error) { return get(string(name), value) })
		if err != nil {
			return nil, err
		}

		xattrs = append(xattrs, place.Xattr{Name: string(name), Value: value})
	}

	return xattrs, nil
}

// sized reads what a call gives into a buffer of the size it asks for.
func sized(call func([]byte) (int, error)) ([]byte, error) {
	size, err := call(nil)
	if err != nil || size == 0 {
		return nil, err
	}

	buffer := make([]byte, size)

	size, err = call(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:size], nil
}
