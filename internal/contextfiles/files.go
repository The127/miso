package contextfiles

import (
	"io"

	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/protocol"
)

// Of are the files of a copy, read from the build context.
func Of(context *buildcontext.Dir) func(protocol.Copy) protocol.Files {
	return func(request protocol.Copy) protocol.Files {
		return func(send func(protocol.Entry, io.Reader) error) error {
			return context.Pack(request.Sources[0], request.Digests[0], func(entry buildcontext.Entry, content io.Reader) error {
				return send(protocol.Entry{Kind: entry.Kind, Path: entry.Path, Mode: entry.Mode, Target: entry.Target, Size: entry.Size}, content)
			})
		}
	}
}
