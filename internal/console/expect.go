package console

import (
	"fmt"
	"regexp"
)

// Expect reads the serial console until the pattern matches and returns
// what came before the match. What came after it is kept for the next
// Expect, so a ^ in the pattern is where the last match ended.
func (c *Console) Expect(pattern *regexp.Regexp) (string, error) {
	chunk := make([]byte, 4096)
	for {
		if match := pattern.FindIndex(c.seen); match != nil {
			before := string(c.seen[:match[0]])
			c.seen = c.seen[match[1]:]

			return before, nil
		}

		if c.end != nil {
			return "", fmt.Errorf("the console ended before %q came: %w", pattern, c.end)
		}

		n, err := c.serial.Read(chunk)
		c.seen = append(c.seen, chunk[:n]...)
		c.end = err
	}
}
