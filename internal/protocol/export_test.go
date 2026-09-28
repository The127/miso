package protocol

// EntriesOf reads the entries of a copy from the host over the connection.
func EntriesOf(c *Conn) Entries {
	return &reader{conn: c}
}
