package tree

import (
	"bytes"

	"golang.org/x/sys/unix"
)

// xattrs gives a copy the extended attributes of what it was copied from,
// never following a link.
func (c copier) xattrs(name string) error {
	return at(c.from, name, func(source string) error {
		return at(c.to, name, func(target string) error {
			attributes, err := list(source)
			if err != nil {
				return err
			}

			for _, attribute := range attributes {
				value, err := get(source, attribute)
				if err != nil {
					return err
				}

				if err := unix.Lsetxattr(target, attribute, value, 0); err != nil {
					return err
				}
			}

			return nil
		})
	})
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

	for _, attribute := range bytes.Split(names[:size], []byte{0}) {
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
