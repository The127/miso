// Package vmtest lets the tests that need root and a real kernel run as the
// init of a small VM. Only tests built with the vmtest tag use it, and
// just test-vm is what boots them.
package vmtest
