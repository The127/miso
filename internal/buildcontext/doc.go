// Package buildcontext reads what a COPY names from the build context, the
// directory on the host that a build file sits in.
//
// # The digest
//
// A digest says what a COPY of a source would put into an image. Two sources
// with the same digest give the same layer. It reads
//
//	miso-context-1:<sha256 in hex>
//
// The name in front stands for everything written here. Whoever changes what
// goes into a digest, or how, changes that name.
//
// A source is walked depth first. Each directory is walked in the byte order
// of its names, and a directory comes before what is in it. A source that is
// a file or a link is a walk of one. Every step of the walk is one entry of
// four fields:
//
//	kind     "file", "directory" or "link"
//	path     as seen from the source, "." for the source itself
//	mode     the twelve unix bits in octal, like 4755, empty for a link
//	payload  for a file the sha256 of its content in hex, for a link its
//	         target as written, empty for a directory
//
// An entry is hashed as its fields in that order. Each field is written as
// its length in bytes in decimal, a colon, and its bytes. The digest is the
// hash of the entry hashes in walk order, each written the same way.
//
// # What stays out
//
// A payload for the builder must carry exactly what a digest covers. What
// stays out of a digest the payload has to drop or set to a constant.
//
//   - Times. A fresh checkout has new ones.
//   - Owners. The host's users mean nothing in an image.
//   - Extended attributes. They carry host policy, like SELinux labels.
//   - Hard links. Two names are two files.
//   - The name of the source. The instruction says it already.
//
// # Links and special files
//
// A link of the build context is never followed on the host, it means a
// place in the image. Inside a tree and as a source it is copied as a link.
// A source path that goes through one is refused.
//
// A pipe, a socket or a device cannot be copied and is refused by name.
package buildcontext
