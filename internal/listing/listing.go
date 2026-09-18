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
	for _, stage := range planned.Stages {
		if _, err := fmt.Fprintf(w, "%s\n", from(stage)); err != nil {
			return err
		}

		for _, step := range stage.Steps {
			line, text := describe(step.Instruction)
			if _, err := fmt.Fprintf(w, "%4d  %s  %s\n", line, short(step.Key), text); err != nil {
				return err
			}

			for _, file := range step.Files {
				if _, err := fmt.Fprintf(w, "%20s%s  %s\n", "", file.Path, short(file.Digest)); err != nil {
					return err
				}
			}
		}
	}

	return nil
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

// short cuts a key to twelve characters, as git does with a commit.
func short(key string) string {
	return key[:12]
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
