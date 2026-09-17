package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Name     string
	Base     string
	Commands []string
}

// Parse reads a build file into its stages, in order.
func Parse(source string) ([]Stage, error) {
	var p parser

	joined := strings.ReplaceAll(source, "\\\n", "")
	for i, line := range strings.Split(joined, "\n") {
		if err := p.read(line); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
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

func (p *parser) read(line string) error {
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	keyword, arguments, _ := strings.Cut(line, " ")
	switch keyword {
	case "FROM":
		return p.from(arguments)
	case "RUN":
		return p.run(arguments)
	default:
		return fmt.Errorf("unknown instruction %s", keyword)
	}
}

func (p *parser) from(arguments string) error {
	words := strings.Fields(arguments)
	named := len(words) == 3 && words[1] == "AS"

	switch {
	case len(words) == 0:
		return errors.New("FROM needs a base")
	case len(words) == 1:
		p.stages = append(p.stages, Stage{Base: words[0]})
	case named:
		p.stages = append(p.stages, Stage{Base: words[0], Name: words[2]})
	default:
		return errors.New("FROM takes a base and an optional AS name")
	}

	return nil
}

func (p *parser) run(command string) error {
	if len(p.stages) == 0 {
		return errors.New("RUN before FROM")
	}

	stage := &p.stages[len(p.stages)-1]
	stage.Commands = append(stage.Commands, command)
	return nil
}
