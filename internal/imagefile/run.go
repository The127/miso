package imagefile

// Run is a shell command, kept verbatim.
type Run struct {
	Command string
}

func (Run) instruction() {}

func (p *parser) run(arguments string) error {
	return p.add("RUN", Run{Command: arguments})
}
