package main

import (
	"slices"

	"github.com/urfave/cli/v3"
)

// takesWords is the metadata of a command whose arguments after -- are the
// words of a command to run. For any other command -- only ends the flags,
// as it does everywhere, and what follows it is as positional as the rest.
const takesWords = "takes-words"

// wordsAfterDashes are the words after the -- that the command was given,
// which urfave drops from its arguments and keeps in those of the root. A
// -- earlier on the line may be the value of a flag, so it is the one that
// leaves exactly the arguments the command has. Only a flag whose value is
// -- and that is followed by nothing but the arguments can fool it.
func wordsAfterDashes(command *cli.Command) []string {
	if command.Metadata[takesWords] != true {
		return nil
	}

	typed := command.Root().Args().Slice()
	args := command.Args().Slice()
	for at, word := range typed {
		if word != "--" {
			continue
		}

		words := typed[at+1:]
		if len(words) <= len(args) && slices.Equal(words, args[len(args)-len(words):]) {
			return words
		}
	}

	return nil
}

// positionalArgs are the arguments of the command that are no words after
// the dashes.
func positionalArgs(command *cli.Command) []string {
	args := command.Args().Slice()

	return args[:len(args)-len(wordsAfterDashes(command))]
}
