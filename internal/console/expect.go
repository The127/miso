package console

import (
	"io"
	"regexp"
)

// Expect reads the serial console until the pattern matches and returns
// what came before the match.
func Expect(serial io.Reader, pattern *regexp.Regexp) (string, error) {
	var seen []byte

	chunk := make([]byte, 64)
	for {
		n, err := serial.Read(chunk)
		seen = append(seen, chunk[:n]...)

		if match := pattern.FindIndex(seen); match != nil {
			return string(seen[:match[0]]), nil
		}

		if err != nil {
			return "", err
		}
	}
}
