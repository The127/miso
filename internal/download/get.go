package download

import (
	"context"
	"fmt"
	"net/http"
)

// Get downloads what a URL stands for and answers the digest of the bytes
// as they arrived.
func (s *Store) Get(ctx context.Context, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	response, err := s.client.Do(request)
	if err != nil {
		return "", err
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, response.Status)
	}

	return s.keep(response.Body)
}
