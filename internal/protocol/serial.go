package protocol

import "strings"

// Serial is the serial the host gives the disk of a base image and the
// agent finds it by. qemu keeps 20 characters of a serial, so it is the
// start of the hash without the name of the hash function.
func Serial(digest string) string {
	hash := digest[strings.Index(digest, ":")+1:]

	return hash[:20]
}
