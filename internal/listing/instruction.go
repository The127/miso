package listing

import (
	"maps"
	"slices"
	"strings"

	"github.com/The127/miso/internal/imagefile"
)

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
