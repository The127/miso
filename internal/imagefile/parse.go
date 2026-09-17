package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Parse reads a build file into its stages, in order.
func Parse(source string) ([]Stage, error) {
	var p parser
	for _, line := range sourceLines(source) {
		if err := p.read(line); err != nil {
			return nil, fmt.Errorf("line %d: %w", line.number, err)
		}
	}

	if len(p.stages) == 0 {
		return nil, errors.New("no FROM instruction")
	}

	return p.stages, nil
}

type parser struct {
	stages []Stage
}

func (p *parser) read(line sourceLine) error {
	if line.problem != nil {
		return line.problem
	}

	text := strings.TrimLeft(line.text, " \t")
	if text == "" || strings.HasPrefix(text, "#") {
		return nil
	}

	keyword, arguments, _ := strings.Cut(text, " ")
	switch strings.ToUpper(keyword) {
	case "FROM":
		return p.from(arguments)
	case "RUN":
		return p.run(arguments)
	case "ENV":
		return p.env(arguments)
	case "COPY":
		return p.copy(arguments)
	default:
		return fmt.Errorf("unknown instruction %s", keyword)
	}
}

func (p *parser) add(keyword string, instruction Instruction) error {
	if len(p.stages) == 0 {
		return fmt.Errorf("%s before FROM", keyword)
	}

	stage := &p.stages[len(p.stages)-1]
	stage.Instructions = append(stage.Instructions, instruction)
	return nil
}
