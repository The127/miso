package buildcontext

// OpenFile lets the tests open a name the way a digest and a pack do,
// without a walk that looked at it first.
var OpenFile = (*Dir).open
