package console

import (
	"context"
	"fmt"
	"regexp"
)

// Expect waits until the pattern matches what the console shows and
// returns what came before the match. What came after it is kept for the
// next Expect, so a ^ in the pattern is where the last match ended. When
// the pattern never comes, it returns what the console showed since the
// last match, so a caller can show why.
func (c *Console) Expect(ctx context.Context, pattern *regexp.Regexp) (string, error) {
	for {
		before, found, end := c.match(pattern)

		switch {
		case found:
			return before, nil
		case end != nil:
			return c.neverCame(pattern, end)
		}

		select {
		case <-c.more:
		case <-ctx.Done():
			return c.neverCame(pattern, ctx.Err())
		}
	}
}

func (c *Console) match(pattern *regexp.Regexp) (string, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if match := pattern.FindIndex(c.seen); match != nil {
		before := string(c.seen[:match[0]])
		c.seen = c.seen[match[1]:]

		return before, true, nil
	}

	return "", false, c.end
}

func (c *Console) neverCame(pattern *regexp.Regexp, cause error) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return string(c.seen), fmt.Errorf("%q never came: %w", pattern, cause)
}
