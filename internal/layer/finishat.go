package layer

import (
	"os"
	"time"
)

// FinishAt puts the layer under its key, marked as used at a time. A build
// marks the layers below at the start of its request and its own layer with
// the same time, so no layer is older than one that stands on it.
func (w *Work) FinishAt(at time.Time) error {
	if err := os.Chtimes(w.dir, at, at); err != nil {
		return err
	}

	return w.Finish()
}
