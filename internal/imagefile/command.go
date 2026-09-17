package imagefile

import (
	"errors"
	"strings"
)

// shellCommand is the text of a RUN or a CHECK. It stays verbatim, but it
// must say something.
func shellCommand(arguments string) (string, error) {
	if strings.TrimSpace(arguments) == "" {
		return "", errors.New("needs a command")
	}

	return arguments, nil
}
