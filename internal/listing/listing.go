package listing

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

// Write lists a plan, a stage as its FROM line and every step as its line
// number, its key and the instruction as written.
func Write(w io.Writer, planned plan.Plan) error {
	out := &lines{w: w}
	if planned.Agent != "" {
		out.print("agent %s", planned.Agent)
	}

	for _, stage := range planned.Stages {
		out.print("%s", from(stage))
		for _, step := range stage.Steps {
			out.step(step)
		}
	}

	return out.err
}

// lines keeps the first error a writer gives, so that the listing can be
// written as its lines and the error is checked once at the end.
type lines struct {
	w   io.Writer
	err error
}

func (l *lines) print(format string, args ...any) {
	if l.err != nil {
		return
	}

	_, l.err = fmt.Fprintf(l.w, format+"\n", args...)
}

func (l *lines) step(step plan.Step) {
	line, text := describe(step.Instruction)
	l.print("%4d  %s  %s", line, short(step.Key), text)

	for _, file := range step.Files {
		l.print("%20s%s  %s", "", file.Path, short(file.Digest))
	}
}

func from(stage plan.Stage) string {
	line := "FROM " + stage.Base
	if stage.Name != "" {
		line += " AS " + stage.Name
	}

	if stage.BaseDigest != "" {
		line += "  base " + short(stage.BaseDigest)
	}

	return line
}

// short cuts a key to twelve characters, as git does with a commit. A
// digest names its format in front, and that stays.
func short(key string) string {
	name, hex, named := strings.Cut(key, ":")
	if !named {
		return key[:12]
	}

	return name + ":" + hex[:12]
}

// describe is an instruction as its author wrote it, and where.
func describe(instruction imagefile.Instruction) (int, string) {
	switch step := instruction.(type) {
	case imagefile.Run:
		return step.Line, "RUN " + step.Command
	case imagefile.Env:
		return step.Line, "ENV " + step.Key + "=" + step.Value
	case imagefile.Copy:
		return step.Line, copyText(step)
	case imagefile.Output:
		return step.Line, outputText(step)
	case imagefile.Check:
		return step.Line, "CHECK " + step.Command
	}

	return 0, ""
}

func copyText(step imagefile.Copy) string {
	words := []string{"COPY"}
	if step.From != "" {
		words = append(words, "--from="+step.From)
	}

	words = append(words, step.Sources...)

	return strings.Join(append(words, step.Destination), " ")
}

// outputText puts the options in one order, a map has none of its own.
func outputText(step imagefile.Output) string {
	words := []string{"OUTPUT", step.Kind, step.Name}
	for _, name := range slices.Sorted(maps.Keys(step.Options)) {
		word := "--" + name
		if value := step.Options[name]; value != "" {
			word += "=" + value
		}

		words = append(words, word)
	}

	return strings.Join(words, " ")
}
