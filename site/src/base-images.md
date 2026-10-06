# Base images

`FROM` starts a stage from a base. miso knows these base images by name:

| Name | Image |
| --- | --- |
| `debian:sid` | The daily Debian sid cloud image. |
| `debian:13` | The current Debian 13 (trixie) cloud image. |
| `debian:trixie` | The same image as `debian:13`. |

The images are the `genericcloud` qcow2 images that Debian publishes at
`cloud.debian.org`, for amd64 and arm64. miso fetches the image for the
architecture of the machine it runs on. `miso build` still accepts only amd64,
see [Command line](cli.md#miso-build). A name that is not in the table is an
error, and so are paths, URLs and container images.

`FROM scratch` starts from an empty root file system, and `FROM <stage>` starts
from an earlier stage of the same file. See [The Imagefile](imagefile.md).

## When an image is fetched

The first build that needs a name downloads the image into the `bases`
directory of the [cache](cli.md#the-cache), or into `bases/arm64` for arm64.
Later builds use the cached image. A name keeps pointing at the same image for
as long as that image stays in the cache, even when Debian has published a
newer one.

`miso plan` shows the image a stage starts from as a digest after the `FROM`
line:

```
FROM debian:sid  base sha256:6a9ef70a8ecb
```

A base image that has not been fetched yet is listed as `download debian:sid`.

## Getting a newer image

The digest of the base image is part of the key of the first step of the
stage. When the image changes, the keys of all the steps of the stage change,
and its layers are built again.

`miso prune` removes downloads that were not used for a while. The next build
that needs a removed image downloads it again, and it may be a newer one. See
[Command line](cli.md#miso-prune).

## Packages

The base images are cloud images. Install what a build needs with `RUN`, for
example `RUN apt-get update && apt-get install -y htop`. A `RUN` is cached by
its text, so the packages it installs are not looked up again on the next
build. See [Caching](caching.md).
