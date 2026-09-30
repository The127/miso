package download

import "time"

// Policy says which blobs are stale. Now is the time it is judged at. A
// limit that is zero is not set, and lets nothing go.
type Policy struct {
	Now       time.Time
	OlderThan time.Duration
}

// Stale gives the digests of the blobs the policy lets go.
func Stale(blobs []Blob, policy Policy) []string {
	var stale []string

	for _, blob := range blobs {
		if policy.OlderThan != 0 && policy.Now.Sub(blob.Used) > policy.OlderThan {
			stale = append(stale, blob.Digest)
		}
	}

	return stale
}
