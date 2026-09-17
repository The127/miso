package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Name         string
	Base         string
	Instructions []Instruction
}

// Instruction is one step of a stage: a Run or an Env.
type Instruction interface {
	instruction()
}

// Run is a shell command, kept verbatim.
type Run struct {
	Command string
}

// Env sets a variable for the instructions after it.
type Env struct {
	Key   string
	Value string
}

func (Run) instruction() {}
func (Env) instruction() {}

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

	text := strings.TrimLeft(line.text, " \t")
	if text == "" || strings.HasPrefix(text, "#") {
		return nil
	}

	keyword, arguments, _ := strings.Cut(text, " ")
	switch keyword {
	case "FROM":
		return p.from(arguments)
	case "RUN":
		return p.add("RUN", Run{Command: arguments})
	case "ENV":
		return p.env(arguments)
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

var errEnvUsage = errors.New("ENV needs KEY=VALUE")

func (p *parser) env(arguments string) error {
	assignments, closed := words(arguments)
	if !closed {
		return errors.New("ENV has an unclosed quote")
	}

	if len(assignments) == 0 {
		return errEnvUsage
	}

	for _, assignment := range assignments {
		key, value, assigned := strings.Cut(assignment, "=")
		if !assigned {
			return errEnvUsage
		}

		if err := p.add("ENV", Env{Key: key, Value: value}); err != nil {
			return err
		}
	}

	return nil
}

// words splits text at spaces and tabs. A double quote keeps the spaces up
// to the next one, the quotes themselves are dropped. It reports whether
// every quote was closed.
func words(text string) (found []string, closed bool) {
	var word strings.Builder
	quoted := false
	flush := func() {
		if word.Len() > 0 {
			found = append(found, word.String())
			word.Reset()
		}
	}

	for _, r := range text {
		switch {
		case r == '"':
			quoted = !quoted
		case !quoted && (r == ' ' || r == '\t'):
			flush()
		default:
			word.WriteRune(r)
		}
	}

	flush()
	return found, !quoted
}

func (p *parser) add(keyword string, instruction Instruction) error {
	if len(p.stages) == 0 {
		return fmt.Errorf("%s before FROM", keyword)
	}

	stage := &p.stages[len(p.stages)-1]
	stage.Instructions = append(stage.Instructions, instruction)
	return nil
}
