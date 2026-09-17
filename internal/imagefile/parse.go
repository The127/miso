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
			return nil, &Error{Line: line.number, Err: err}
		}
	}

	if len(p.stages) == 0 {
		return nil, errors.New("no FROM instruction")
	}

	return p.stages, nil
}

// reader turns the arguments of one instruction into what goes into the
// stage. Its errors read on from the keyword, which the parser puts in
// front: "needs a base" becomes "FROM needs a base".
type reader func(line int, arguments string) ([]Instruction, error)

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

	written, arguments, _ := strings.Cut(text, " ")
	keyword := strings.ToUpper(written)

	var read reader
	switch keyword {
	case "FROM":
		return p.open(line.number, arguments)
	case "RUN":
		read = readRun
	case "ENV":
		read = readEnv
	case "COPY":
		read = readCopy
	case "OUTPUT":
		read = readOutput
	default:
		return fmt.Errorf("%w %s", ErrUnknownInstruction, written)
	}

	if len(p.stages) == 0 {
		return fmt.Errorf("%s %w", keyword, ErrBeforeFrom)
	}

	instructions, err := read(line.number, arguments)
	if err != nil {
		return fmt.Errorf("%s %w", keyword, err)
	}

	stage := &p.stages[len(p.stages)-1]
	stage.Instructions = append(stage.Instructions, instructions...)
	return nil
}

func (p *parser) open(line int, arguments string) error {
	stage, err := readFrom(line, arguments)
	if err != nil {
		return fmt.Errorf("FROM %w", err)
	}

	p.stages = append(p.stages, stage)
	return nil
}
