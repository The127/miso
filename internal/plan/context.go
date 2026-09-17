package plan

// Context is the directory a COPY takes its files from.
type Context interface {
	Digest(path string) string
}
