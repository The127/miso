package layer

// Rename is Finish stopped before its last sync, where a crash must still
// find the layer whole.
func Rename(w *Work) error {
	return w.rename()
}
