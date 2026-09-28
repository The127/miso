package copydigest

import "strconv"

// Entry is one step of the walk of a source, all that a digest may know of
// it besides a file's content.
type Entry struct {
	// "file", "directory" or "link"
	Kind string

	// as seen from the source, "." for the source itself
	Path string

	// the twelve unix bits, none for a link
	Mode uint32

	// of a link, as written
	Target string
}

// Sum is the hash of an entry. It takes the hash of a file's content, so
// that a payload can stream the content first and sum the entry after.
func (e Entry) Sum(content string) string {
	payload := e.Target
	if e.Kind == "file" {
		payload = content
	}

	// written in octal, like 4755, and empty for a link, which has no mode
	// of its own
	mode := strconv.FormatUint(uint64(e.Mode), 8)
	if e.Kind == "link" {
		mode = ""
	}

	return hashed([]string{e.Kind, e.Path, mode, payload})
}
