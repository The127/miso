package agent

import "time"

// found marks the layers below a key as used, whether a build needs them
// or a hit skips them, and tells whether the layer of the key is there. If
// it is, that is marked too. All get the same time, which is returned for a
// build to finish its layer with, so no layer is older than one that stands
// on it.
func (a *Agent) found(key string, below []string) (bool, time.Time, error) {
	now := time.Now()

	if err := a.layers.Use(now, below...); err != nil {
		return false, now, err
	}

	there, err := a.layers.Has(key)
	if err != nil || !there {
		return false, now, err
	}

	return true, now, a.layers.Use(now, key)
}
