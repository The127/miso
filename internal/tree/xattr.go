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
	var attributes []attribute

	err := at(c.from, "getxattr", name, func(source string) error {
		var err error

		attributes, err = read(source)

		return err
	})
	if err != nil {
		return err
	}

	return at(c.to, "setxattr", name, func(target string) error {
		for _, a := range attributes {
			if err := unix.Lsetxattr(target, a.name, a.value, 0); err != nil {
				return err
			}
		}

		return nil
	})
}

type attribute struct {
	name  string
	value []byte
}

// read is every extended attribute of a path with its value.
func read(path string) ([]attribute, error) {
	names, err := list(path)
	if err != nil {
		return nil, err
	}

	attributes := make([]attribute, 0, len(names))

	for _, name := range names {
		value, err := get(path, name)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes, attribute{name: name, value: value})
	}

	return attributes, nil
}

// list names the extended attributes of a path.
func list(path string) ([]string, error) {
	size, err := unix.Llistxattr(path, nil)
	if err != nil || size == 0 {
		return nil, err
	}

	names := make([]byte, size)

	size, err = unix.Llistxattr(path, names)
	if err != nil {
		return nil, err
	}

	var attributes []string

	for attribute := range bytes.SplitSeq(names[:size], []byte{0}) {
		if len(attribute) > 0 {
			attributes = append(attributes, string(attribute))
		}
	}

	return attributes, nil
}

// get is the value of an extended attribute of a path.
func get(path, attribute string) ([]byte, error) {
	size, err := unix.Lgetxattr(path, attribute, nil)
	if err != nil {
		return nil, err
	}

	value := make([]byte, size)

	size, err = unix.Lgetxattr(path, attribute, value)
	if err != nil {
		return nil, err
	}

	return value[:size], nil
}

// fileXattrs is every extended attribute of an open file with its value.
func fileXattrs(file *os.File) ([]place.Xattr, error) {
	fd := int(file.Fd())

	size, err := unix.Flistxattr(fd, nil)
	if err != nil || size == 0 {
		return nil, err
	}

	names := make([]byte, size)

	size, err = unix.Flistxattr(fd, names)
	if err != nil {
		return nil, err
	}

	var xattrs []place.Xattr

	for name := range bytes.SplitSeq(names[:size], []byte{0}) {
		if len(name) == 0 {
			continue
		}

		value, err := fileXattr(fd, string(name))
		if err != nil {
			return nil, err
		}

		xattrs = append(xattrs, place.Xattr{Name: string(name), Value: value})
	}

	return xattrs, nil
}

func fileXattr(fd int, name string) ([]byte, error) {
	size, err := unix.Fgetxattr(fd, name, nil)
	if err != nil {
		return nil, err
	}

	value := make([]byte, size)

	size, err = unix.Fgetxattr(fd, name, value)
	if err != nil {
		return nil, err
	}

	return value[:size], nil
}
