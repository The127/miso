// Package buildcontext reads what a COPY names from the build context, the
// directory on the host that a build file sits in.
//
// What a COPY carries and what its digest covers is the copydigest's. This
// package reads it from the host.
//
// # Links and special files
//
// A link of the build context is never followed on the host, it means a
// place in the image. Inside a tree and as a source it is copied as a link.
// A source path that goes through one is refused.
//
// A pipe, a socket or a device cannot be copied and is refused by name.
package buildcontext
