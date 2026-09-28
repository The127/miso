package protocol

// Entry is one file, directory or link of a copy, as the digest of the
// build context covers it. A file's content follows as raw bytes.
type Entry struct {
	// the position of its source among the sources of the copy
	Source int

	// "file", "directory" or "link"
	Kind string

	// as seen from the source, "." for the source itself
	Path string

	// the twelve unix bits, none for a link
	Mode uint32

	// of a link, as written
	Target string

	// how many raw bytes of content follow, none but for a file
	Size int64
}

func (e Entry) into(envelope *envelope) { envelope.Entry = &e }
