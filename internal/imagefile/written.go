package imagefile

import (
	"maps"
	"slices"
	"strings"
)

// Written is an instruction as its author wrote it, and where.
func Written(instruction Instruction) (int, string) {
	switch step := instruction.(type) {
	case Run:
		if step.Offline {
			return step.Line, "RUN --network=none " + step.Command
		}

		return step.Line, "RUN " + step.Command
	case Env:
		return step.Line, "ENV " + step.Key + "=" + step.Value
	case Copy:
		return step.Line, copyText(step)
	case Output:
		return step.Line, outputText(step)
	case Check:
		return step.Line, "CHECK " + step.Command
	case Partition:
		return step.Line, partitionText(step)
	case Cmdline:
		return step.Line, "CMDLINE " + step.Text
	}

	return 0, ""
}

func copyText(step Copy) string {
	words := []string{"COPY"}
	if step.From != "" {
		words = append(words, "--from="+step.From)
	}

	words = append(words, step.Sources...)

	return strings.Join(append(words, step.Destination), " ")
}

// outputText puts the options in one order, a map has none of its own.
func outputText(step Output) string {
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

func partitionText(step Partition) string {
	words := []string{"PARTITION", step.Name}
	for _, setting := range step.Settings {
		words = append(words, setting.Key+"="+setting.Value)
	}

	return strings.Join(words, " ")
}
