package baseimage

import (
	"context"
	"fmt"
	"net/http"
)

// Fetch downloads the image a name stands for and answers its digest, that
// of the bytes as they arrived.
func (c *Cache) Fetch(ctx context.Context, name string) (string, error) {
	url, err := c.source(name)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	response, err := c.client.Do(request)
	if err != nil {
		return "", err
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, response.Status)
	}

	digest, err := c.store(response.Body)
	if err != nil {
		return "", err
	}

	return digest, c.remember(name, digest)
}
