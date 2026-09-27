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

	// before the name, so a name never points at an image of no format
	if err := c.record(digest, source.Format); err != nil {
		return "", err
	}

	return digest, c.remember(name, digest)
}
