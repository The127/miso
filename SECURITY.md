# Security Policy

## Reporting a vulnerability

Please report vulnerabilities privately via
[GitHub private vulnerability reporting](https://github.com/The127/miso/security/advisories/new).
Do not open a public issue for security problems.

You can expect a timely response. Please include enough
detail to reproduce the issue (affected component, setup, steps).

There is no bug bounty program.

## Supported versions

miso is in early development and has no releases yet. Only the `main`
branch is supported. Once versioned releases exist, this section will state
which release lines receive security fixes.

## Scope notes

miso builds operating system images from a build file. A build file is
trusted input: its `RUN` steps execute as root inside the builder VM, with
network access, and that is the job. Of particular interest are issues in:

- the boundary between the builder VM and the host (a build step, a base
  image or a build context reaching the host beyond the output directory,
  a `COPY` source escaping the build context),
- the integrity of what goes into an image (base image downloads, a cached
  layer served for inputs it was not built from),
- the artifacts themselves (a verity root hash, kernel command line or
  partition table that does not match the image, build-time secrets such
  as keys or machine identity left in a layer or an output),
- the stock installer (writing to a disk other than the chosen target).
