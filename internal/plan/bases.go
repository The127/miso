package plan

// Bases knows the images a FROM can start a stage on.
type Bases interface {
	Digest(base string) string
}

// root is the key a stage starts from. The agent is part of it because a
// changed agent may build the same step differently.
func root(agent string, base string, bases Bases) string {
	fields := []string{agent, base}
	if base != "scratch" {
		fields = append(fields, bases.Digest(base))
	}

	return hashed(fields)
}
