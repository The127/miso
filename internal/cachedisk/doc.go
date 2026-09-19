// Package cachedisk makes the disk the builder keeps its layers on, a file
// on the host that outlives every builder VM, and locks it for one build at
// a time.
package cachedisk
