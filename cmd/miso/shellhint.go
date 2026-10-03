package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
)

// withShellHint is the error of a step with a line added that says how to
// get a shell before that step, on the same build file. A FROM is no step,
// so a failed import gets none.
func withShellHint(command *cli.Command, requests []build.Request, err error) error {
	located, isLocated := errors.AsType[*imagefile.Error](err)
	if !isLocated || located.Line == 0 || build.IsFrom(requests, located.Line) {
		return err
	}

	words := []string{"miso", "shell", "--before", strconv.Itoa(located.Line)}
	if file := command.String(fileFlagName); strings.HasPrefix(file, "-") {
		words = append(words, "--file="+file)
	} else if file != "" {
		words = append(words, "-f", file)
	}

	if positional := positionalArgs(command); len(positional) > 0 {
		words = append(words, notAFlag(positional[0]))
	}

	hinted := *located
	hinted.Err = fmt.Errorf("%w\nto look at the layers before it: %s", located.Err, quoted(words))

	return &hinted
}

// quoted are the words as a shell takes them back, which only those with
// something a shell would read differently need quotes.
func quoted(words []string) string {
	quotedWords := make([]string, len(words))
	for i, word := range words {
		quotedWords[i] = word
		if word == "" || strings.ContainsAny(word, " \t\n\"'\\$`&|;<>()*?[]{}#~!") {
			quotedWords[i] = "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
		}
	}

	return strings.Join(quotedWords, " ")
}

// notAFlag is the path with a dot in front of a dash, which would be read as
// a flag.
func notAFlag(path string) string {
	if strings.HasPrefix(path, "-") {
		return "./" + path
	}

	return path
}
