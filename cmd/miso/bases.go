package main

import "errors"

// noBases stands in until miso can fetch base images.
type noBases struct{}

func (noBases) Digest(string) (string, error) {
	return "", errors.New("no base images yet, a FROM takes scratch or an earlier stage")
}
