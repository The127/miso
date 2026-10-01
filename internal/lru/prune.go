package lru

// Swept is what a sweep or a prune removed.
type Swept struct {
	Count int
	Bytes int64
}

// Prune removes the items the policy lets go through remove, and says how
// much went. It stops at the first item that cannot be removed.
func Prune(items []Item, policy Policy, remove func(id string) error) (Swept, error) {
	var swept Swept

	sizes := make(map[string]int64, len(items))
	for _, item := range items {
		sizes[item.ID] = item.Size
	}

	for _, id := range Stale(items, policy) {
		if err := remove(id); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += sizes[id]
	}

	return swept, nil
}
