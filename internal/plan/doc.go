// Package plan turns the stages of a build file into a plan, before anything
// is built. It checks that the stages make sense together, looks up what
// they take from outside the build file, and gives every step its cache key.
package plan
