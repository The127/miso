package console

import (
	"fmt"
	"io"
)

// Send types the text on the console.
func (c *Console) Send(text string) error {
	if _, err := io.WriteString(c.keyboard, text); err != nil {
		return fmt.Errorf("typing on the console: %w", err)
	}

	return nil
}
