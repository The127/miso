package tree

// landed is where each file with several names first landed in the image.
// A place something else took since holds that file no more, so a later
// name must not link to it.
type landed struct {
	first map[inode]string
	at    map[string]inode
}

func newLanded() landed {
	return landed{first: map[inode]string{}, at: map[string]inode{}}
}

// where is where a file first landed, while it is still there.
func (l landed) where(file inode) (string, bool) {
	target, ok := l.first[file]

	return target, ok
}

// put has a file first land at a target.
func (l landed) put(file inode, target string) {
	l.first[file] = target
	l.at[target] = file
}

// taken forgets the file that first landed at a target something else now
// takes.
func (l landed) taken(target string) {
	if file, ok := l.at[target]; ok {
		delete(l.first, file)
		delete(l.at, target)
	}
}
