package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// splitOptions separates the words that start with -- from the rest. An
// option is --name or --name=value.
func splitOptions(words []string) (rest []string, options map[string]string, err error) {
	for _, word := range words {
		option, isOption := strings.CutPrefix(word, "--")
		if !isOption {
			rest = append(rest, word)
			continue
		}

		name, value, _ := strings.Cut(option, "=")
		if name == "" {
			return nil, nil, errors.New("has an option without a name")
		}

		if _, twice := options[name]; twice {
			return nil, nil, fmt.Errorf("has --%s twice", name)
		}

		if options == nil {
			options = map[string]string{}
		}

		options[name] = value
	}

	return rest, options, nil
}
