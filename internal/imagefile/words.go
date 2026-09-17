package imagefile

import (
	"errors"
	"strings"
)

// words splits text at spaces and tabs. A double quote keeps the spaces up
// to the next one, the quotes themselves are dropped.
func words(text string) ([]string, error) {
	var found []string
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
	if quoted {
		return nil, errors.New("has an unclosed quote")
	}

	return found, nil
}
