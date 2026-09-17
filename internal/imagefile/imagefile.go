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

// sourceLine is a line with its continuations joined in, numbered by where
// it starts in the file. It is unfinished when its last continuation has
// no line to continue on.
type sourceLine struct {
	number     int
	text       string
	unfinished bool
}

func sourceLines(source string) []sourceLine {
	var lines []sourceLine
	var current sourceLine
	for i, physical := range strings.Split(strings.TrimSuffix(source, "\n"), "\n") {
		if current.number == 0 {
			current.number = i + 1
		}

		head, continued := strings.CutSuffix(physical, "\\")
		current.text += head
		if continued {
			continue
		}

		lines = append(lines, current)
		current = sourceLine{}
	}

	if current.number != 0 {
		current.unfinished = true
		lines = append(lines, current)
	}

	return lines
}

type parser struct {
	stages []Stage
}

func (p *parser) read(line sourceLine) error {
	if line.unfinished {
		return errors.New("continuation runs off the end of the file")
	}

	if line.text == "" || strings.HasPrefix(line.text, "#") {
		return nil
	}

	keyword, arguments, _ := strings.Cut(line.text, " ")
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
