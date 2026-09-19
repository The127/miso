package sandbox

// mounts are what the helper mounts in the root of a run, in order, each on
// a directory of the floor.
var mounts = []struct {
	point string
	mount func() error
}{
	{"proc", mountProc},
	{"sys", mountSys},
	{"dev", mountDev},
}
