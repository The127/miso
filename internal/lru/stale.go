package lru

import (
	"slices"
	"time"
)

// Item is something the cache holds. Used is when it was last used, and
// Size the bytes it takes.
type Item struct {
	ID   string
	Used time.Time
	Size int64
}

// Policy says which items are stale. Now is the time it is judged at. A
// limit that is zero is not set, and lets nothing go. KeepStorage is the
// bytes the items that stay may take.
type Policy struct {
	Now         time.Time
	OlderThan   time.Duration
	KeepStorage int64
}

// Stale gives the IDs of the items the policy lets go.
func Stale(items []Item, policy Policy) []string {
	var stale []string
	var kept []Item

	for _, item := range items {
		if policy.OlderThan != 0 && policy.Now.Sub(item.Used) > policy.OlderThan {
			stale = append(stale, item.ID)

			continue
		}

		kept = append(kept, item)
	}

	if policy.KeepStorage == 0 {
		return stale
	}

	slices.SortStableFunc(kept, func(a, b Item) int { return a.Used.Compare(b.Used) })

	var total int64
	for _, item := range kept {
		total += item.Size
	}

	// items used together are a chain, as layers are, and a chain that is
	// cut in the middle leaves one whose parent is gone
	for i := 0; i < len(kept) && total > policy.KeepStorage; {
		at := kept[i].Used

		for ; i < len(kept) && kept[i].Used.Equal(at); i++ {
			stale = append(stale, kept[i].ID)
			total -= kept[i].Size
		}
	}

	return stale
}
