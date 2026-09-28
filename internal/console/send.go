package console

import "io"

// Send types the text on the console.
func (c *Console) Send(text string) error {
	_, err := io.WriteString(c.keyboard, text)

	return err
}
