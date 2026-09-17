package plan

// Bases knows the images a FROM can start a stage on.
type Bases interface {
	Digest(base string) string
}
