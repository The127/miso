package layer

// Rename is Finish stopped before its last sync, where a crash must still
// find the layer whole.
func Rename(w *Work) error {
	return w.rename()
}

// Retire is Remove stopped before its files go, where a crash must leave
// nothing under the key.
func Retire(s *Store, key string) (string, error) {
	return s.retire(key)
}
