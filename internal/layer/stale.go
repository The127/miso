package layer

import (
	"slices"
	"time"
)

// Policy says which layers are stale. Now is the time it is judged at. A
// limit that is zero is not set, and lets nothing go. KeepStorage is the
// bytes the layers that stay may take.
type Policy struct {
	Now         time.Time
	OlderThan   time.Duration
	KeepStorage int64
}

// Stale gives the keys of the entries the policy lets go.
func Stale(entries []Entry, policy Policy) []string {
	var stale []string
	var kept []Entry

	for _, entry := range entries {
		if policy.OlderThan != 0 && policy.Now.Sub(entry.Used) > policy.OlderThan {
			stale = append(stale, entry.Key)

			continue
		}

		kept = append(kept, entry)
	}

	if policy.KeepStorage == 0 {
		return stale
	}

	slices.SortStableFunc(kept, func(a, b Entry) int { return a.Used.Compare(b.Used) })

	var total int64
	for _, entry := range kept {
		total += entry.Size
	}

	// layers used together are a chain, and a chain that is cut in the
	// middle leaves a layer whose parent is gone
	for i := 0; i < len(kept) && total > policy.KeepStorage; {
		at := kept[i].Used

		for ; i < len(kept) && kept[i].Used.Equal(at); i++ {
			stale = append(stale, kept[i].Key)
			total -= kept[i].Size
		}
	}

	return stale
}
