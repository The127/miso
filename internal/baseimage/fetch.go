package baseimage

import "context"

// Fetch downloads the image a name stands for and answers its digest, that
// of the bytes as they arrived.
func (c *Cache) Fetch(ctx context.Context, name string) (string, error) {
	source, err := c.source(name)
	if err != nil {
		return "", err
	}

	digest, err := c.blobs.Get(ctx, source.URL)
	if err != nil {
		return "", err
	}

	c.formats[digest] = source.Format

	return digest, c.remember(name, digest)
}

// Format is that of the image with a digest, as its source declared it.
func (c *Cache) Format(digest string) (string, error) {
	return c.formats[digest], nil
}
