package layer

import "time"

// Policy says which layers are stale. Now is the time it is judged at. A
// limit that is zero is not set, and lets nothing go.
type Policy struct {
	Now       time.Time
	OlderThan time.Duration
}

// Stale gives the keys of the entries the policy lets go.
func Stale(entries []Entry, policy Policy) []string {
	var stale []string

	if policy.OlderThan == 0 {
		return stale
	}

	for _, entry := range entries {
		if policy.Now.Sub(entry.Used) > policy.OlderThan {
			stale = append(stale, entry.Key)
		}
	}

	return stale
}
